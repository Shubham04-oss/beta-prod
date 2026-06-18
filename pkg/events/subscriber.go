package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

// Subscriber defines an interface for listening to domain events
type Subscriber interface {
	Subscribe(ctx context.Context, subscriptionID string, handler func(ctx context.Context, event DomainEvent) error) error
	Close() error
}

type PubSubSubscriber struct {
	client *pubsub.Client
}

func NewPubSubSubscriber(ctx context.Context, projectID string) (*PubSubSubscriber, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}
	return &PubSubSubscriber{client: client}, nil
}

// Subscribe listens to a Pub/Sub subscription and unmarshals messages into DomainEvents.
func (s *PubSubSubscriber) Subscribe(ctx context.Context, subscriptionID string, handler func(ctx context.Context, event DomainEvent) error) error {
	sub := s.client.Subscription(subscriptionID)
	
	log.Printf("Starting listener on subscription: %s", subscriptionID)
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var domainEvent DomainEvent
		if err := json.Unmarshal(msg.Data, &domainEvent); err != nil {
			log.Printf("Failed to unmarshal domain event: %v", err)
			msg.Nack()
			return
		}

		// Execute the business logic handler
		if err := handler(ctx, domainEvent); err != nil {
			log.Printf("Handler failed for event %s: %v", domainEvent.EventID, err)
			msg.Nack()
			return
		}

		// Acknowledge the message upon success
		msg.Ack()
	})

	return err
}

func (s *PubSubSubscriber) Close() error {
	return s.client.Close()
}
