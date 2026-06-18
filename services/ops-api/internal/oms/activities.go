package oms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
	"go.temporal.io/sdk/temporal"
)

type UnifiedChannelService interface {
	SyncOrderStatusToChannel(ctx context.Context, orderID, tenantIDStr, orgIDStr string) error
}

type Activities struct {
	repo    *Repository
	unified UnifiedChannelService
}

func NewActivities(repo *Repository, services ...UnifiedChannelService) *Activities {
	var uSvc UnifiedChannelService
	if len(services) > 0 {
		uSvc = services[0]
	}
	return &Activities{repo: repo, unified: uSvc}
}

type LineItemReq struct {
	VariantID        *string `json:"variant_id"`
	LocationID       *string `json:"location_id"`
	SKU              *string `json:"sku"`
	ProductTitle     string  `json:"product_title"`
	VariantTitle     *string `json:"variant_title"`
	UnitPrice        float64 `json:"unit_price"`
	Quantity         int     `json:"quantity"`
	RequiresShipping *bool   `json:"requires_shipping"`
}

type CreateOrderRequest struct {
	TenantID         string        `json:"tenant_id"`
	OrgID            string        `json:"org_id"`
	UserID           string        `json:"user_id"`
	Role             string        `json:"role"`
	CustomerID       *string       `json:"customer_id"`
	IdempotencyKey   *string       `json:"idempotency_key"`
	Currency         *string       `json:"currency"`
	PaymentProvider  *string       `json:"payment_provider"`
	PaymentReference *string       `json:"payment_reference"`
	Channel          *string       `json:"channel"`
	SourcePlatform   *string       `json:"source_platform"`
	Items            []LineItemReq `json:"items"`
}

func (a *Activities) CreateOrderActivity(ctx context.Context, req CreateOrderRequest) (string, error) {
	if err := validateCreateOrderRequest(req); err != nil {
		return "", err
	}

	tenantID := req.TenantID
	orgID := req.OrgID
	userID := req.UserID
	role := req.Role
	ctx = authcontext.WithTenantID(ctx, tenantID)
	ctx = authcontext.WithOrgID(ctx, orgID)
	if userID != "" {
		ctx = authcontext.WithUserID(ctx, userID)
	}
	if role != "" {
		ctx = authcontext.WithRole(ctx, role)
	}

	var orderIDStr string

	err := a.repo.WithTx(ctx, func(tx pgx.Tx, qtx db.Querier) error {
		orderID := uuid.New()
		orderIDStr = orderID.String()

		// Calculate total
		total := 0.0
		for _, item := range req.Items {
			total += item.UnitPrice * float64(item.Quantity)
		}

		var custID pgtype.UUID
		if req.CustomerID != nil {
			id, err := uuid.Parse(*req.CustomerID)
			if err != nil {
				return err
			}
			custID.Bytes = id
			custID.Valid = true
		}

		var idempotencyKey pgtype.Text
		if req.IdempotencyKey != nil {
			idempotencyKey.String = *req.IdempotencyKey
			idempotencyKey.Valid = true
		}

		// Insert Order Draft (Now PENDING)
		_, err := tx.Exec(ctx, `
			INSERT INTO orders (
				id, org_id, tenant_id, customer_id, status, currency,
				subtotal, total, idempotency_key, payment_provider, payment_reference,
				channel, source_platform
			)
			VALUES ($1, $2, $3, $4, 'pending_payment', COALESCE(NULLIF($5, ''), 'USD'), $6, $7, $8, $9, $10, $11, $12)
		`, orderID, orgID, tenantID, custID, textOrEmpty(req.Currency), total, total, idempotencyKey, nullableText(req.PaymentProvider), nullableText(req.PaymentReference), nullableText(req.Channel), nullableText(req.SourcePlatform))
		if err != nil {
			return err
		}

		// Insert items
		for _, item := range req.Items {
			var variantID pgtype.UUID
			if item.VariantID != nil {
				id, err := uuid.Parse(*item.VariantID)
				if err != nil {
					return err
				}
				variantID.Bytes = id
				variantID.Valid = true
			}

			var sku pgtype.Text
			if item.SKU != nil {
				sku.String = *item.SKU
				sku.Valid = true
			}

			var varTitle pgtype.Text
			if item.VariantTitle != nil {
				varTitle.String = *item.VariantTitle
				varTitle.Valid = true
			}

			_, err := tx.Exec(ctx, `
				INSERT INTO order_line_items (org_id, tenant_id, order_id, variant_id, sku, product_title, variant_title, unit_price, quantity, line_total)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, orgID, tenantID, orderID, variantID, sku, item.ProductTitle, varTitle, item.UnitPrice, item.Quantity, item.UnitPrice*float64(item.Quantity))
			if err != nil {
				return err
			}
		}

		payloadBytes, _ := json.Marshal(req)
		actorID := nullableUUID(&userID)

		// Audit Log
		_, err = tx.Exec(ctx, `
			INSERT INTO order_events (org_id, tenant_id, order_id, event_type, actor_id, actor_role, payload)
			VALUES ($1, $2, $3, 'order.created', $4, $5, $6)
		`, orgID, tenantID, orderID, actorID, role, payloadBytes)
		if err != nil {
			return err
		}

		metadataBytes, _ := json.Marshal(map[string]string{
			"actor_id":   userID,
			"actor_role": role,
		})

		// Outbox
		_, err = tx.Exec(ctx, `
			INSERT INTO oms_outbox_events (topic, aggregate_id, aggregate_type, tenant_id, org_id, payload, metadata)
			VALUES ('order.created', $1, 'order', $2, $3, $4, $5)
		`, orderID, tenantID, orgID, payloadBytes, metadataBytes)
		return err
	})

	if err != nil {
		return "", err
	}

	return orderIDStr, nil
}

func validateCreateOrderRequest(req CreateOrderRequest) error {
	if req.TenantID == "" || req.OrgID == "" {
		return fmt.Errorf("tenant_id and org_id are required")
	}
	if req.IdempotencyKey == nil || *req.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("at least one order item is required")
	}
	for i, item := range req.Items {
		if item.ProductTitle == "" {
			return fmt.Errorf("items[%d].product_title is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("items[%d].quantity must be greater than zero", i)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("items[%d].unit_price cannot be negative", i)
		}
		if item.VariantID != nil && *item.VariantID != "" {
			if _, err := uuid.Parse(*item.VariantID); err != nil {
				return fmt.Errorf("items[%d].variant_id must be a uuid", i)
			}
			if item.LocationID == nil || *item.LocationID == "" {
				return fmt.Errorf("items[%d].location_id is required for inventory-backed items", i)
			}
			if _, err := uuid.Parse(*item.LocationID); err != nil {
				return fmt.Errorf("items[%d].location_id must be a uuid", i)
			}
		}
	}
	return nil
}

type LineItemReservation struct {
	VariantID  string
	LocationID string
	Quantity   int
}

func (a *Activities) ReserveInventoryActivity(ctx context.Context, orderID string, items []LineItemReservation) error {
	tenantID, _ := authcontext.GetTenantID(ctx)
	orgID, _ := authcontext.GetOrgID(ctx)

	err := a.repo.WithTx(ctx, func(tx pgx.Tx, qtx db.Querier) error {
		for _, item := range items {
			varID, err1 := uuid.Parse(item.VariantID)
			locID, err2 := uuid.Parse(item.LocationID)
			tenantUUID, err3 := uuid.Parse(tenantID)
			orgUUID, err4 := uuid.Parse(orgID)
			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				return fmt.Errorf("invalid uuid")
			}

			_, err := qtx.ReserveInventoryForOrder(ctx, db.ReserveInventoryForOrderParams{
				ReservedQuantity: int32(item.Quantity),
				VariantID:        pgtype.UUID{Bytes: varID, Valid: true},
				TenantID:         pgtype.UUID{Bytes: tenantUUID, Valid: true},
				LocationID:       pgtype.UUID{Bytes: locID, Valid: true},
				OrgID:            pgtype.UUID{Bytes: orgUUID, Valid: true},
			})
			if err != nil {
				return err // Rollback on any failure
			}
		}
		return nil
	})

	if err != nil {
		return temporal.NewApplicationError("insufficient stock: "+err.Error(), "InsufficientStock", err)
	}
	return nil
}

func (a *Activities) ReleaseInventoryActivity(ctx context.Context, orderID string, items []LineItemReservation) error {
	tenantID, _ := authcontext.GetTenantID(ctx)
	orgID, _ := authcontext.GetOrgID(ctx)

	return a.repo.WithTx(ctx, func(tx pgx.Tx, qtx db.Querier) error {
		for _, item := range items {
			varID, err1 := uuid.Parse(item.VariantID)
			locID, err2 := uuid.Parse(item.LocationID)
			tenantUUID, err3 := uuid.Parse(tenantID)
			orgUUID, err4 := uuid.Parse(orgID)
			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				return fmt.Errorf("invalid uuid")
			}

			err := qtx.ReleaseInventoryForOrder(ctx, db.ReleaseInventoryForOrderParams{
				ReservedQuantity: int32(item.Quantity),
				VariantID:        pgtype.UUID{Bytes: varID, Valid: true},
				TenantID:         pgtype.UUID{Bytes: tenantUUID, Valid: true},
				LocationID:       pgtype.UUID{Bytes: locID, Valid: true},
				OrgID:            pgtype.UUID{Bytes: orgUUID, Valid: true},
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *Activities) AuthorizePaymentActivity(ctx context.Context, orderID string, req CreateOrderRequest) error {
	if req.PaymentProvider == nil || *req.PaymentProvider == "" || req.PaymentReference == nil || *req.PaymentReference == "" {
		return temporal.NewNonRetryableApplicationError("payment_provider and payment_reference are required before order confirmation", "PaymentAuthorizationRequired", nil)
	}

	tenantID, err := authcontext.GetTenantID(ctx)
	if err != nil || tenantID == "" {
		tenantID = req.TenantID
	}
	ctx = authcontext.WithTenantID(ctx, tenantID)
	if req.OrgID != "" {
		ctx = authcontext.WithOrgID(ctx, req.OrgID)
	}
	if req.UserID != "" {
		ctx = authcontext.WithUserID(ctx, req.UserID)
	}
	if req.Role != "" {
		ctx = authcontext.WithRole(ctx, req.Role)
	}

	return a.repo.WithTx(ctx, func(tx pgx.Tx, qtx db.Querier) error {
		ordUUID, err1 := uuid.Parse(orderID)
		tenantUUID, err2 := uuid.Parse(tenantID)
		orgUUID, err3 := uuid.Parse(req.OrgID)
		if err1 != nil || err2 != nil || err3 != nil {
			return fmt.Errorf("invalid uuid")
		}

		if err := qtx.UpdateOrderPaymentStatus(ctx, db.UpdateOrderPaymentStatusParams{
			PaymentStatus:    pgtype.Text{String: "authorized", Valid: true},
			PaymentProvider:  pgtype.Text{String: *req.PaymentProvider, Valid: true},
			PaymentReference: pgtype.Text{String: *req.PaymentReference, Valid: true},
			ID:               pgtype.UUID{Bytes: ordUUID, Valid: true},
			TenantID:         pgtype.UUID{Bytes: tenantUUID, Valid: true},
			OrgID:            pgtype.UUID{Bytes: orgUUID, Valid: true},
		}); err != nil {
			return err
		}

		return qtx.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
			Status:   db.OrderStatusPaymentAuthorized,
			ID:       pgtype.UUID{Bytes: ordUUID, Valid: true},
			TenantID: pgtype.UUID{Bytes: tenantUUID, Valid: true},
			OrgID:    pgtype.UUID{Bytes: orgUUID, Valid: true},
		})
	})
}

func (a *Activities) ConfirmOrderActivity(ctx context.Context, orderID string, req CreateOrderRequest) error {
	tenantID, _ := authcontext.GetTenantID(ctx)
	if tenantID == "" {
		tenantID = req.TenantID
	}
	orgID, _ := authcontext.GetOrgID(ctx)
	if orgID == "" {
		orgID = req.OrgID
	}
	userID, _ := authcontext.GetUserID(ctx)
	if userID == "" {
		userID = req.UserID
	}
	role, _ := authcontext.GetRole(ctx)
	if role == "" {
		role = req.Role
	}
	ctx = authcontext.WithTenantID(ctx, tenantID)
	if orgID != "" {
		ctx = authcontext.WithOrgID(ctx, orgID)
	}
	if userID != "" {
		ctx = authcontext.WithUserID(ctx, userID)
	}
	if role != "" {
		ctx = authcontext.WithRole(ctx, role)
	}

	return a.repo.WithTx(ctx, func(tx pgx.Tx, qtx db.Querier) error {
		ordUUID, err1 := uuid.Parse(orderID)
		tenantUUID, err2 := uuid.Parse(tenantID)
		orgUUID, err3 := uuid.Parse(orgID)
		if err1 != nil || err2 != nil || err3 != nil {
			return fmt.Errorf("invalid uuid")
		}

		err := qtx.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
			Status:   db.OrderStatusConfirmed,
			ID:       pgtype.UUID{Bytes: ordUUID, Valid: true},
			TenantID: pgtype.UUID{Bytes: tenantUUID, Valid: true},
			OrgID:    pgtype.UUID{Bytes: orgUUID, Valid: true},
		})
		if err != nil {
			return err
		}

		payloadBytes := []byte(`{"status":"confirmed"}`)

		actorID := nullableUUID(&userID)
		metadataBytes, _ := json.Marshal(map[string]string{
			"actor_id":   userID,
			"actor_role": role,
		})

		_, err = tx.Exec(ctx, `
			INSERT INTO order_events (org_id, tenant_id, order_id, event_type, actor_id, actor_role, payload)
			VALUES ($1, $2, $3, 'order.confirmed', $4, $5, $6)
		`, orgID, tenantID, orderID, actorID, role, payloadBytes)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO oms_outbox_events (topic, aggregate_id, aggregate_type, tenant_id, org_id, payload, metadata)
			VALUES ('order.confirmed', $1, 'order', $2, $3, $4, $5)
		`, orderID, tenantID, orgID, payloadBytes, metadataBytes)

		return err
	})
}

func (a *Activities) EmitOrderPlacedActivity(ctx context.Context, orderID string) error {
	// TRUE NO-OP: We no longer need an activity to flush the outbox to Pub/Sub.
	// We are using the Postgres LISTEN/NOTIFY trigger on oms_outbox_events.
	// As soon as ConfirmOrderActivity commits, the event is immediately streamed!
	return nil
}

func (a *Activities) MarkOrderFulfilledActivity(ctx context.Context, orderID string) error {
	tenantID, orgID, userID, role, err := workflowIdentity(ctx)
	if err != nil {
		return err
	}
	return a.repo.WithTx(ctx, func(tx pgx.Tx, qtx db.Querier) error {
		ordUUID, tenantUUID, orgUUID, err := parseScopedOrder(orderID, tenantID, orgID)
		if err != nil {
			return err
		}

		var hasShippedFulfillment bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM fulfillments
				WHERE order_id = $1
					AND tenant_id = $2
					AND org_id = $3
					AND status IN ('shipped', 'delivered')
					AND cancelled_at IS NULL
			)
		`, ordUUID, tenantUUID, orgUUID).Scan(&hasShippedFulfillment); err != nil {
			return err
		}
		if !hasShippedFulfillment {
			return temporal.NewNonRetryableApplicationError("no shipped fulfillment found for order", "FulfillmentNotReady", nil)
		}

		if err := qtx.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
			Status:   db.OrderStatusFulfilled,
			ID:       pgtype.UUID{Bytes: ordUUID, Valid: true},
			TenantID: pgtype.UUID{Bytes: tenantUUID, Valid: true},
			OrgID:    pgtype.UUID{Bytes: orgUUID, Valid: true},
		}); err != nil {
			return err
		}

		payloadBytes := []byte(`{"status":"fulfilled"}`)
		metadataBytes, _ := json.Marshal(map[string]string{"actor_id": userID, "actor_role": role})
		actorID := nullableUUID(&userID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_events (org_id, tenant_id, order_id, event_type, actor_id, actor_role, payload)
			VALUES ($1, $2, $3, 'order.fulfilled', $4, $5, $6)
		`, orgID, tenantID, orderID, actorID, role, payloadBytes); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO oms_outbox_events (topic, aggregate_id, aggregate_type, tenant_id, org_id, payload, metadata)
			VALUES ('order.fulfilled', $1, 'order', $2, $3, $4, $5)
		`, orderID, tenantID, orgID, payloadBytes, metadataBytes)
		return err
	})
}

func (a *Activities) MarkReturnRequestedActivity(ctx context.Context, orderID string) error {
	tenantID, orgID, userID, role, err := workflowIdentity(ctx)
	if err != nil {
		return err
	}
	return a.repo.WithTx(ctx, func(tx pgx.Tx, qtx db.Querier) error {
		ordUUID, tenantUUID, orgUUID, err := parseScopedOrder(orderID, tenantID, orgID)
		if err != nil {
			return err
		}

		var hasOpenReturn bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM returns
				WHERE order_id = $1
					AND tenant_id = $2
					AND org_id = $3
					AND status IN ('requested', 'authorized', 'in_transit', 'received', 'inspected', 'restocked')
			)
		`, ordUUID, tenantUUID, orgUUID).Scan(&hasOpenReturn); err != nil {
			return err
		}
		if !hasOpenReturn {
			return temporal.NewNonRetryableApplicationError("no open return found for order", "ReturnNotFound", nil)
		}

		if err := qtx.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
			Status:   db.OrderStatusReturnRequested,
			ID:       pgtype.UUID{Bytes: ordUUID, Valid: true},
			TenantID: pgtype.UUID{Bytes: tenantUUID, Valid: true},
			OrgID:    pgtype.UUID{Bytes: orgUUID, Valid: true},
		}); err != nil {
			return err
		}

		payloadBytes := []byte(`{"status":"return_requested"}`)
		metadataBytes, _ := json.Marshal(map[string]string{"actor_id": userID, "actor_role": role})
		actorID := nullableUUID(&userID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_events (org_id, tenant_id, order_id, event_type, actor_id, actor_role, payload)
			VALUES ($1, $2, $3, 'order.return_requested', $4, $5, $6)
		`, orgID, tenantID, orderID, actorID, role, payloadBytes); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO oms_outbox_events (topic, aggregate_id, aggregate_type, tenant_id, org_id, payload, metadata)
			VALUES ('order.return_requested', $1, 'order', $2, $3, $4, $5)
		`, orderID, tenantID, orgID, payloadBytes, metadataBytes)
		return err
	})
}

func (a *Activities) SyncFulfillmentToChannelActivity(ctx context.Context, orderID string) error {
	if a.unified == nil {
		return nil
	}
	tenantID, _ := authcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil // No tenant ID means it might be a system test
	}
	orgID, _ := authcontext.GetOrgID(ctx)
	if orgID == "" {
		return nil
	}

	err := a.unified.SyncOrderStatusToChannel(ctx, orderID, tenantID, orgID)
	if err != nil {
		return err
	}
	return nil
}

func workflowIdentity(ctx context.Context) (string, string, string, string, error) {
	tenantID, err := authcontext.GetTenantID(ctx)
	if err != nil || tenantID == "" {
		return "", "", "", "", temporal.NewNonRetryableApplicationError("tenant id required", "MissingTenant", err)
	}
	orgID, err := authcontext.GetOrgID(ctx)
	if err != nil || orgID == "" {
		return "", "", "", "", temporal.NewNonRetryableApplicationError("org id required", "MissingOrg", err)
	}
	userID, err := authcontext.GetUserID(ctx)
	if err != nil || userID == "" {
		userID = uuid.Nil.String()
	}
	role, err := authcontext.GetRole(ctx)
	if err != nil || role == "" {
		role = "SYSTEM"
	}
	return tenantID, orgID, userID, role, nil
}

func parseScopedOrder(orderID, tenantID, orgID string) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	ordUUID, err1 := uuid.Parse(orderID)
	tenantUUID, err2 := uuid.Parse(tenantID)
	orgUUID, err3 := uuid.Parse(orgID)
	if err1 != nil || err2 != nil || err3 != nil {
		return uuid.UUID{}, uuid.UUID{}, uuid.UUID{}, temporal.NewNonRetryableApplicationError("invalid scoped order identity", "InvalidScopedOrder", nil)
	}
	return ordUUID, tenantUUID, orgUUID, nil
}

func nullableText(value *string) pgtype.Text {
	if value == nil || *value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func textOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nullableUUID(value *string) pgtype.UUID {
	if value == nil || *value == "" {
		return pgtype.UUID{}
	}
	parsed, err := uuid.Parse(*value)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}
}
