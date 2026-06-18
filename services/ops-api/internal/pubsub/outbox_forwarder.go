package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/events"
)

type OutboxForwarder struct {
	pool   *pgxpool.Pool
	client *pubsub.Client
}

func NewOutboxForwarder(pool *pgxpool.Pool, client *pubsub.Client) *OutboxForwarder {
	return &OutboxForwarder{
		pool:   pool,
		client: client,
	}
}

func (f *OutboxForwarder) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := f.listenAndForward(ctx)
				if err != nil {
					log.Printf("OutboxForwarder error: %v, retrying in 5 seconds...", err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()
}

func (f *OutboxForwarder) listenAndForward(ctx context.Context) error {
	conn, err := f.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN outbox_insert")
	if err != nil {
		return err
	}

	// Do an initial flush just in case we missed events while restarting
	if err := f.flushOutbox(ctx); err != nil {
		log.Printf("Error flushing outbox initially: %v", err)
	}

	log.Println("OutboxForwarder listening for 'outbox_insert' notifications...")

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}

		log.Printf("Received outbox notification: %s", notification.Payload)
		if err := f.flushOutbox(ctx); err != nil {
			log.Printf("Error flushing outbox: %v", err)
		}
	}
}

func (f *OutboxForwarder) flushOutbox(ctx context.Context) error {
	rows, err := f.pool.Query(ctx, `
		SELECT id::text, topic, aggregate_id::text, aggregate_type, tenant_id::text, org_id::text, payload, COALESCE(metadata, '{}'::jsonb), created_at
		FROM oms_outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT 100
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var toUpdate []string

	for rows.Next() {
		var id string
		var topic string
		var aggregateID string
		var aggregateType string
		var tenantID string
		var orgID string
		var payload json.RawMessage
		var metadata json.RawMessage
		var createdAt time.Time

		if err := rows.Scan(&id, &topic, &aggregateID, &aggregateType, &tenantID, &orgID, &payload, &metadata, &createdAt); err != nil {
			log.Printf("Error scanning outbox row: %v", err)
			continue
		}
		meta := map[string]string{}
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &meta); err != nil {
				log.Printf("Error decoding outbox metadata for %s: %v", id, err)
				continue
			}
		}
		userID := meta["actor_id"]
		if userID == "" {
			userID = "00000000-0000-0000-0000-000000000000"
		}
		role := meta["actor_role"]
		if role == "" {
			role = "SYSTEM"
		}

		event := events.DomainEvent{
			EventID:   id,
			EventType: topic,
			Timestamp: createdAt.UTC(),
			TenantID:  tenantID,
			OrgID:     orgID,
			UserID:    userID,
			Role:      role,
			Payload:   payload,
		}
		eventBytes, err := json.Marshal(event)
		if err != nil {
			log.Printf("Error encoding outbox event %s: %v", id, err)
			continue
		}

		t := f.client.Topic(topic)
		result := t.Publish(ctx, &pubsub.Message{
			Data: eventBytes,
			Attributes: map[string]string{
				"event_id":       id,
				"event_type":     topic,
				"tenant_id":      tenantID,
				"org_id":         orgID,
				"user_id":        userID,
				"role":           role,
				"aggregate_id":   aggregateID,
				"aggregate_type": aggregateType,
			},
		})

		// Block until published
		_, err = result.Get(ctx)
		if err != nil {
			log.Printf("Failed to publish outbox event %s to topic %s: %v", id, topic, err)
			continue
		}

		toUpdate = append(toUpdate, id)
	}

	if len(toUpdate) > 0 {
		_, err := f.pool.Exec(ctx, `
			UPDATE oms_outbox_events
			SET published_at = now()
			WHERE id = ANY($1)
		`, toUpdate)
		if err != nil {
			log.Printf("Error marking outbox events as published: %v", err)
		} else {
			log.Printf("Successfully published and marked %d events", len(toUpdate))
		}
	}

	return nil
}
