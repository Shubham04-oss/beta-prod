package oms

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type OrderCreationParams struct {
	Request      CreateOrderRequest
	Reservations []LineItemReservation
}

func OrderCreationWorkflow(ctx workflow.Context, params OrderCreationParams) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Temporal Propagator extracts tenant_id, org_id, etc. automatically
	// into the workflow context, and will inject them back when calling activities.

	var a *Activities
	var orderID string

	// Step 1: Create Order Draft
	err := workflow.ExecuteActivity(ctx, a.CreateOrderActivity, params.Request).Get(ctx, &orderID)
	if err != nil {
		return "", err
	}

	// Step 2: Reserve Inventory
	err = workflow.ExecuteActivity(ctx, a.ReserveInventoryActivity, orderID, params.Reservations).Get(ctx, nil)
	if err != nil {
		// Cannot reserve inventory, cancel order
		// Optional: we could trigger CancelOrderActivity here to mark it failed.
		return "", err
	}

	// Step 3: Authorize Payment
	err = workflow.ExecuteActivity(ctx, a.AuthorizePaymentActivity, orderID, params.Request).Get(ctx, nil)
	if err != nil {
		// Compensation: release inventory
		workflow.ExecuteActivity(ctx, a.ReleaseInventoryActivity, orderID, params.Reservations).Get(ctx, nil)
		return "", err
	}

	// Step 4: Confirm Order
	err = workflow.ExecuteActivity(ctx, a.ConfirmOrderActivity, orderID, params.Request).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	// Step 5: Emit Events
	workflow.ExecuteActivity(ctx, a.EmitOrderPlacedActivity, orderID).Get(ctx, nil)

	return orderID, nil
}

// OrderFulfillmentWorkflow orchestrates the fulfillment steps
func OrderFulfillmentWorkflow(ctx workflow.Context, orderID string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var a *Activities
	if err := workflow.ExecuteActivity(ctx, a.MarkOrderFulfilledActivity, orderID).Get(ctx, nil); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, a.SyncFulfillmentToChannelActivity, orderID).Get(ctx, nil)
}

// OrderReturnWorkflow orchestrates the returns
func OrderReturnWorkflow(ctx workflow.Context, orderID string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var a *Activities
	return workflow.ExecuteActivity(ctx, a.MarkReturnRequestedActivity, orderID).Get(ctx, nil)
}
