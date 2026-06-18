package unified

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/synq/ops-api/internal/oms"
	"github.com/synq/pkg/authcontext"
	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

// InboundOrderPollingWorkflow runs on a cron schedule
func InboundOrderPollingWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, "PollCommerceOrdersActivity").Get(ctx, nil)
	return err
}

func (a *Activities) PollCommerceOrdersActivity(ctx context.Context) error {
	// 1. Fetch all active connections
	env := os.Getenv("UNIFIED_ENV")
	if env == "" {
		return fmt.Errorf("UNIFIED_ENV is required for inbound order polling")
	}
	res, err := a.unifiedSDK.SDK().Unified.ListUnifiedConnections(ctx, operations.ListUnifiedConnectionsRequest{
		Env: &env,
	})
	if err != nil {
		return fmt.Errorf("failed to list unified connections: %w", err)
	}

	for _, conn := range res.Connections {
		if conn.ID == nil || *conn.ID == "" {
			continue
		}
		// Log the polling attempt
		log.Printf("Polling orders for connection %s (%s)", *conn.ID, conn.IntegrationType)

		// 2. Query Accounting Orders
		ordersRes, err := a.unifiedSDK.SDK().Accounting.ListAccountingOrders(ctx, operations.ListAccountingOrdersRequest{
			ConnectionID: *conn.ID,
		})
		if err != nil {
			log.Printf("Failed to poll orders for connection %s: %v", *conn.ID, err)
			continue
		}

		// 3. For each order, we would start the child workflow
		for _, order := range ordersRes.AccountingOrders {
			if order.ID == nil || *order.ID == "" {
				continue
			}
			// Verify connection maps to a tenant (we do this in the workflow too, but good to check early)
			var orgIDStr string
			var tenantIDStr string
			err := a.dbpool.QueryRow(ctx, `
				SELECT org_id::text, tenant_id::text
				FROM commerce_connections
				WHERE unified_connection_id = $1 AND status = 'ACTIVE' AND deleted_at IS NULL
			`, *conn.ID).Scan(&orgIDStr, &tenantIDStr)
			if err != nil {
				log.Printf("Unmapped connection %s, skipping order", *conn.ID)
				continue
			}

			// Marshal and Unmarshal into map
			b, _ := json.Marshal(order)
			var dataMap map[string]interface{}
			_ = json.Unmarshal(b, &dataMap)

			payload := UnifiedWebhookPayload{
				WorkspaceID:  "default",
				ConnectionID: *conn.ID,
				Event:        "accounting.order.created",
				Data:         dataMap,
			}

			// For simplicity in the activity, we use the temporal client to trigger the workflow.
			workflowCtx := authcontext.WithTenantID(ctx, tenantIDStr)
			workflowCtx = authcontext.WithOrgID(workflowCtx, orgIDStr)
			workflowCtx = authcontext.WithUserID(workflowCtx, uuid.Nil.String())
			workflowCtx = authcontext.WithRole(workflowCtx, "SYSTEM")
			_, err = a.temporalClient.ExecuteWorkflow(workflowCtx, client.StartWorkflowOptions{
				ID:        fmt.Sprintf("sync-inbound-%s-%s", *conn.ID, *order.ID),
				TaskQueue: "oms-task-queue-v3",
			}, SyncInboundOrderWorkflow, payload)

			if err != nil {
				log.Printf("Failed to start sync workflow for order %s: %v", *order.ID, err)
			}
		}
	}
	return nil
}

// SyncInboundOrderWorkflow processes a single inbound order
func SyncInboundOrderWorkflow(ctx workflow.Context, payload UnifiedWebhookPayload) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Map Unified AccountingOrder to OMS CreateOrderRequest
	var req oms.CreateOrderRequest
	err := workflow.ExecuteActivity(ctx, "MapInboundOrderActivity", payload).Get(ctx, &req)
	if err != nil {
		return err
	}

	// Step 2: Call OMS OrderCreationWorkflow as a child workflow
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("oms-create-order-unified-%s", *req.IdempotencyKey),
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	params := oms.OrderCreationParams{
		Request:      req,
		Reservations: []oms.LineItemReservation{}, // Will be calculated inside
	}

	err = workflow.ExecuteChildWorkflow(ctx, oms.OrderCreationWorkflow, params).Get(ctx, nil)
	return err
}

func (a *Activities) MapInboundOrderActivity(ctx context.Context, payload UnifiedWebhookPayload) (oms.CreateOrderRequest, error) {
	var orgID string
	var tenantID string
	var provider string
	if err := a.dbpool.QueryRow(ctx, `
		SELECT org_id::text, tenant_id::text, provider
		FROM commerce_connections
		WHERE unified_connection_id = $1 AND status = 'ACTIVE' AND deleted_at IS NULL
		LIMIT 1
	`, payload.ConnectionID).Scan(&orgID, &tenantID, &provider); err != nil {
		return oms.CreateOrderRequest{}, fmt.Errorf("unmapped unified connection %s: %w", payload.ConnectionID, err)
	}

	externalOrderID := extractString(payload.Data, "id")
	if externalOrderID == "" {
		externalOrderID = uuid.NewString()
	}
	idempotencyKey := fmt.Sprintf("unified:%s:%s", payload.ConnectionID, externalOrderID)

	items := []oms.LineItemReq{}

	if rawItems, ok := payload.Data["line_items"].([]interface{}); ok {
		for _, ri := range rawItems {
			if itemMap, ok := ri.(map[string]interface{}); ok {
				var variantIDPtr *string
				if idStr, ok := itemMap["item_id"].(string); ok && idStr != "" {
					variantIDPtr = &idStr
				}
				var skuPtr *string
				if sku, ok := itemMap["sku"].(string); ok && sku != "" {
					skuPtr = &sku
				}
				qty := 1
				if q, ok := itemMap["quantity"].(float64); ok {
					qty = int(q)
				}
				title := extractString(itemMap, "name")
				if title == "" {
					title = extractString(itemMap, "description")
				}
				if title == "" {
					title = "Unified order item"
				}
				unitPrice := 0.0
				if amount, ok := itemMap["unit_amount"].(float64); ok {
					unitPrice = amount
				} else if amount, ok := itemMap["amount"].(float64); ok {
					unitPrice = amount
				}
				items = append(items, oms.LineItemReq{
					VariantID:    variantIDPtr,
					SKU:          skuPtr,
					ProductTitle: title,
					Quantity:     qty,
					UnitPrice:    unitPrice,
				})
			}
		}
	}
	if len(items) == 0 {
		return oms.CreateOrderRequest{}, fmt.Errorf("unified order %s has no line items", externalOrderID)
	}

	currency := extractString(payload.Data, "currency")
	if currency == "" {
		currency = "USD"
	}
	paymentProvider := provider
	paymentReference := externalOrderID
	sourcePlatform := provider

	req := oms.CreateOrderRequest{
		TenantID:         tenantID,
		OrgID:            orgID,
		IdempotencyKey:   &idempotencyKey,
		Currency:         &currency,
		PaymentProvider:  &paymentProvider,
		PaymentReference: &paymentReference,
		SourcePlatform:   &sourcePlatform,
		Items:            items,
	}
	return req, nil
}

func extractString(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}
