package ucp

import (
	"context"
	"log"

	"github.com/synq/pkg/events"
	"go.temporal.io/sdk/client"
)

// StartPIMEventSubscriber initializes the event-driven UCP feed generation.
func StartPIMEventSubscriber(ctx context.Context, sub events.Subscriber, temporalClient client.Client, subscriptionID string) error {
	
	handler := func(ctx context.Context, event events.DomainEvent) error {
		// Only trigger UCP regeneration for PIM mutation events
		switch event.EventType {
		case "synq.pim.product.created",
			"synq.pim.product.updated",
			"synq.pim.product.deleted",
			"synq.pim.variant.created",
			"synq.pim.variant.updated",
			"synq.pim.media.created",
			"synq.pim.media.deleted":
			
			// Extract the TenantID directly from the 4 Pillars imprinted on the event
			tenantID := event.TenantID
			
			log.Printf("Received %s. Triggering UCP Feed Workflow for Tenant: %s", event.EventType, tenantID)

			// Kick off the Temporal Workflow to regenerate the UCP feed asynchronously
			options := client.StartWorkflowOptions{
				ID:        "ucp_feed_generator_" + tenantID,
				TaskQueue: "UCP_TASK_QUEUE",
			}
			we, err := temporalClient.ExecuteWorkflow(ctx, options, GenerateUCPFeedWorkflow, tenantID)
			if err != nil {
				return err
			}
			log.Printf("Started Temporal workflow WorkflowID: %s, RunID: %s", we.GetID(), we.GetRunID())
			
		default:
			// Ignore other events
		}
		
		return nil
	}

	// This is a blocking call that listens indefinitely
	return sub.Subscribe(ctx, subscriptionID, handler)
}
