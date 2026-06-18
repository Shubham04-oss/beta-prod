package unified

import (
	"context"
	"encoding/json"
	"log"

	"github.com/synq/pkg/events"
)

// StartUnifiedEventSubscriber listens to PIM events and triggers the worker pool for outbox sync
func StartUnifiedEventSubscriber(ctx context.Context, sub events.Subscriber, service *Service, subscriptionID string) error {

	handler := func(ctx context.Context, event events.DomainEvent) error {
		switch event.EventType {
		case "synq.pim.product.created", "synq.pim.product.updated":

			var payloadData map[string]interface{}
			if err := json.Unmarshal(event.Payload, &payloadData); err != nil {
				return nil
			}
			productID, ok := payloadData["id"].(string)
			if !ok {
				return nil
			}
			log.Printf("[Unified Sync] Enqueuing UPSERT for Product: %s", productID)
			service.EnqueuePushJob(event.TenantID, event.OrgID, productID, "UPSERT")

		case "synq.pim.product.deleted":

			var payloadData map[string]interface{}
			if err := json.Unmarshal(event.Payload, &payloadData); err != nil {
				return nil
			}
			productID, ok := payloadData["id"].(string)
			if !ok {
				return nil
			}
			log.Printf("[Unified Sync] Enqueuing DELETE for Product: %s", productID)
			service.EnqueuePushJob(event.TenantID, event.OrgID, productID, "DELETE")

		case "synq.pim.variant.created", "synq.pim.variant.updated", "synq.pim.variant.deleted",
			"synq.pim.media.created", "synq.pim.media.updated", "synq.pim.media.deleted":

			var payloadData map[string]interface{}
			if err := json.Unmarshal(event.Payload, &payloadData); err != nil {
				return nil
			}
			productID, ok := payloadData["product_id"].(string)
			if !ok {
				return nil
			}
			// When child changes, UPSERT the parent product
			log.Printf("[Unified Sync] Child event %s. Enqueuing UPSERT for Parent Product: %s", event.EventType, productID)
			service.EnqueuePushJob(event.TenantID, event.OrgID, productID, "UPSERT")

		default:
			// Ignore other events
		}

		return nil
	}

	// Blocks and listens indefinitely
	return sub.Subscribe(ctx, subscriptionID, handler)
}
