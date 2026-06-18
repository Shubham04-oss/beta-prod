package procurement

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/db"
	"github.com/synq/pkg/events"
)

type Service interface {
	ProcessInventoryEvent(ctx context.Context, event events.DomainEvent) error
}

type service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) Service {
	return &service{
		pool: pool,
	}
}

// StartProcurementEventSubscriber listens to the event bus for inventory adjustments
func StartProcurementEventSubscriber(ctx context.Context, sub events.Subscriber, svc Service, subscriptionID string) error {
	log.Printf("Starting Procurement listener on subscription: %s", subscriptionID)
	err := sub.Subscribe(ctx, subscriptionID, func(ctx context.Context, event events.DomainEvent) error {
		return svc.ProcessInventoryEvent(ctx, event)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	return nil
}

type InventoryAdjustedPayload struct {
	VariantID  string `json:"variant_id"`
	TenantID   string `json:"tenant_id"`
	OrgID      string `json:"org_id"`
	LocationID string `json:"location_id"`
}

func (s *service) ProcessInventoryEvent(ctx context.Context, event events.DomainEvent) error {
	// Ensure this is the right event type
	if event.EventType != "synq.pim.inventory.adjusted" {
		return nil
	}

	var payload InventoryAdjustedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal inventory event: %v", err)
	}

	if payload.TenantID == "" || payload.VariantID == "" {
		return nil // Ignore invalid payloads
	}

	var tenantID, orgID, variantID pgtype.UUID
	tenantID.Scan(payload.TenantID)
	orgID.Scan(payload.OrgID)
	variantID.Scan(payload.VariantID)

	queries := db.New(s.pool)

	// 1. Fetch Tenant Settings
	settings, err := queries.GetTenantSettings(ctx, db.GetTenantSettingsParams{
		TenantID: tenantID,
		OrgID:    orgID,
	})
	if err != nil {
		// If no settings exist, default to auto_po disabled
		return nil
	}

	if !settings.AutoPoEnabled.Valid || !settings.AutoPoEnabled.Bool {
		return nil // Auto PO disabled
	}

	threshold := int32(10) // Default
	if settings.DefaultLowStockThreshold.Valid {
		threshold = settings.DefaultLowStockThreshold.Int32
	}

	// 2. Fetch current inventory level across all locations for this variant
	// (Assuming we check total stock, not just location specific)
	var totalAvailable int32
	err = s.pool.QueryRow(ctx, "SELECT COALESCE(SUM(available_quantity), 0) FROM inventory_levels WHERE tenant_id = $1 AND org_id = $2 AND variant_id = $3", tenantID, orgID, variantID).Scan(&totalAvailable)
	if err != nil {
		return err
	}

	if totalAvailable >= threshold {
		return nil // Stock is healthy
	}

	// 3. Draft a Smart Purchase Order
	log.Printf("Low stock detected for variant %s (Qty: %d, Threshold: %d). Generating Smart PO...", variantID.Bytes, totalAvailable, threshold)

	// Get or Create a default supplier
	supplier, err := queries.GetFirstSupplier(ctx, tenantID)
	if err != nil {
		// Create a dummy supplier for this tenant
		supplier, err = queries.CreateSupplier(ctx, db.CreateSupplierParams{
			OrgID:    orgID,
			TenantID: tenantID,
			Name:     "Default Supplier (Auto-Generated)",
		})
		if err != nil {
			return fmt.Errorf("failed to get/create supplier: %v", err)
		}
	}

	// Create a new PO
	po, err := queries.CreatePurchaseOrder(ctx, db.CreatePurchaseOrderParams{
		OrgID:      orgID,
		TenantID:   tenantID,
		SupplierID: supplier.ID,
		Status:     db.NullPoStatus{PoStatus: db.PoStatusDRAFT, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create draft PO: %v", err)
	}

	// Calculate suggested order quantity
	suggestedQuantity := threshold - totalAvailable + 10 // Order enough to pass threshold + buffer

	// Create PO Line Item
	_, err = queries.CreatePurchaseOrderLineItem(ctx, db.CreatePurchaseOrderLineItemParams{
		OrgID:     orgID,
		TenantID:  tenantID,
		PoID:      po.ID,
		VariantID: variantID,
		Quantity:  suggestedQuantity,
		UnitPrice: pgtype.Numeric{}, // We could grab from variant.cost_price
		Subtotal:  pgtype.Numeric{},
	})
	if err != nil {
		return fmt.Errorf("failed to create draft PO line item: %v", err)
	}

	log.Printf("Successfully generated Draft PO %s for Variant %s", po.ID.Bytes, variantID.Bytes)
	return nil
}
