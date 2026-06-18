package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/synq/pkg/authcontext"
)

// DomainEvent represents a standard structure for all events published by the system.
// It inherently carries the 4 Pillars of Identity (TenantID, OrgID, UserID, Role)
// for strict event-driven isolation.
type DomainEvent struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	TenantID  string          `json:"tenant_id"`
	OrgID     string          `json:"org_id"`
	UserID    string          `json:"user_id"`
	Role      string          `json:"role"`
	Payload   json.RawMessage `json:"payload"`
}

// Publisher is an interface for publishing domain events.
type Publisher interface {
	Publish(ctx context.Context, topicID, eventType string, payload interface{}) error
	Close() error
}

// PubSubPublisher implements Publisher using Google Cloud Pub/Sub.
type PubSubPublisher struct {
	client *pubsub.Client
}

// NewPubSubPublisher creates a new PubSubPublisher.
func NewPubSubPublisher(ctx context.Context, projectID string) (*PubSubPublisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}
	return &PubSubPublisher{client: client}, nil
}

// Publish enriches the payload with the 4 Identity Pillars from the context and publishes it.
func (p *PubSubPublisher) Publish(ctx context.Context, topicID, eventType string, payload interface{}) error {
	// Extract the 4 IDs securely from the context
	tenantID, err := authcontext.GetTenantID(ctx)
	if err != nil {
		return err
	}
	orgID, err := authcontext.GetOrgID(ctx)
	if err != nil {
		return err
	}
	userID, err := authcontext.GetUserID(ctx)
	if err != nil {
		return err
	}
	role, err := authcontext.GetRole(ctx)
	if err != nil {
		return err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := DomainEvent{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		TenantID:  tenantID,
		OrgID:     orgID,
		UserID:    userID,
		Role:      role,
		Payload:   payloadBytes,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	topic := p.client.Topic(topicID)

	// Publish the message asynchronously and wait for the result
	result := topic.Publish(ctx, &pubsub.Message{
		Data: eventBytes,
		Attributes: map[string]string{
			"event_type": eventType,
			"tenant_id":  tenantID,
			"org_id":     orgID,
		},
	})

	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topicID, err)
	}

	return nil
}

func (p *PubSubPublisher) Close() error {
	return p.client.Close()
}
