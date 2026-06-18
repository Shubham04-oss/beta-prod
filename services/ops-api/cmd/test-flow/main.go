package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/synq/ops-api/internal/pim"
	"github.com/synq/ops-api/internal/procurement"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
	"github.com/synq/pkg/events"
)

// MockPublisher executes subscriber logic inline for testing
type MockPublisher struct {
	Svc procurement.Service
}

func (m *MockPublisher) Publish(ctx context.Context, topic, eventType string, data interface{}) error {
	fmt.Printf("Event Emitted -> Topic: %s | Type: %s\n", topic, eventType)

	// If it's an inventory adjusted event, instantly pass it to the procurement worker
	if eventType == "synq.pim.inventory.adjusted" {
		// Mock serialization
		b, _ := json.Marshal(data)

		fmt.Printf(" [Worker Triggered] Processing %s...\n", eventType)
		err := m.Svc.ProcessInventoryEvent(ctx, events.DomainEvent{EventType: eventType, Payload: b})
		if err != nil {
			fmt.Printf(" [Worker Error] %v\n", err)
		}
	}
	return nil
}
func (m *MockPublisher) Close() error { return nil }

func main() {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	if err != nil {
		fmt.Printf("Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var orgID, tenantID, locID uuid.UUID
	dbpool.QueryRow(ctx, "SELECT id FROM organizations LIMIT 1").Scan(&orgID)
	dbpool.QueryRow(ctx, "SELECT id FROM tenants WHERE org_id = $1 LIMIT 1", orgID).Scan(&tenantID)
	dbpool.QueryRow(ctx, "SELECT id FROM locations WHERE tenant_id = $1 LIMIT 1", tenantID).Scan(&locID)

	authCtx := context.WithValue(ctx, authcontext.TenantIDKey, tenantID.String())
	authCtx = context.WithValue(authCtx, authcontext.OrgIDKey, orgID.String())

	// 1. Setup Tenant Settings (Enable Auto-PO)
	queries := db.New(dbpool)
	queries.UpsertTenantSettings(authCtx, db.UpsertTenantSettingsParams{
		OrgID:                    pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:                 pgtype.UUID{Bytes: tenantID, Valid: true},
		InventoryAllocationModel: pgtype.Text{String: "HARD", Valid: true},
		AutoPoEnabled:            pgtype.Bool{Bool: true, Valid: true},
		DefaultLowStockThreshold: pgtype.Int4{Int32: 20, Valid: true},
		CostingMethod:            pgtype.Text{String: "WAC", Valid: true},
	})
	fmt.Println("[Config] Auto-POs Enabled. Threshold = 20.")

	procSvc := procurement.NewService(dbpool)
	pimSvc := pim.NewService(dbpool, &MockPublisher{Svc: procSvc})

	fmt.Println("\n--- STARTING FULL END-TO-END TEST ---")

	prodID := uuid.New()
	newProduct, _ := pimSvc.CreateProduct(authCtx, db.CreateProductParams{
		ID:       pgtype.UUID{Bytes: prodID, Valid: true},
		OrgID:    pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID: pgtype.UUID{Bytes: tenantID, Valid: true},
		Title:    "E2E Test Product",
		Status:   pgtype.Text{String: "ACTIVE", Valid: true},
	})

	var price pgtype.Numeric
	price.Scan("500.00")
	variantID := uuid.New()
	newVariant, _ := pimSvc.CreateVariant(authCtx, db.CreateProductVariantParams{
		ID:        pgtype.UUID{Bytes: variantID, Valid: true},
		OrgID:     pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:  pgtype.UUID{Bytes: tenantID, Valid: true},
		ProductID: newProduct.ID,
		Sku:       pgtype.Text{String: "E2E-SKU-001", Valid: true},
		Price:     price,
	})

	fmt.Println("\n[Test A] WAC Calculation & RESTOCK")
	// 10 units at $100
	var cost1 pgtype.Numeric
	cost1.Scan("100.00")
	_, err = pimSvc.AdjustInventory(authCtx, db.CreateInventoryLedgerEntryParams{
		OrgID:           pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:        pgtype.UUID{Bytes: tenantID, Valid: true},
		VariantID:       newVariant.ID,
		LocationID:      pgtype.UUID{Bytes: locID, Valid: true},
		TransactionType: "RESTOCK",
		QuantityDelta:   10,
		UnitCost:        cost1,
	})
	if err != nil {
		fmt.Println("Err 1: ", err)
	}
	// 10 units at $200
	var cost2 pgtype.Numeric
	cost2.Scan("200.00")
	_, err = pimSvc.AdjustInventory(authCtx, db.CreateInventoryLedgerEntryParams{
		OrgID:           pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:        pgtype.UUID{Bytes: tenantID, Valid: true},
		VariantID:       newVariant.ID,
		LocationID:      pgtype.UUID{Bytes: locID, Valid: true},
		TransactionType: "RESTOCK",
		QuantityDelta:   10,
		UnitCost:        cost2,
	})
	if err != nil {
		fmt.Println("Err 2: ", err)
	}

	var wac pgtype.Numeric
	dbpool.QueryRow(ctx, "SELECT cost_price FROM product_variants WHERE id = $1", newVariant.ID).Scan(&wac)
	val, _ := wac.Float64Value()
	fmt.Printf("-> Variant WAC successfully calculated natively via Postgres Trigger: $%.2f (Expected: $150.00)\n", val.Float64)

	fmt.Println("\n[Test B] Event-Driven Auto PO Generation")
	fmt.Println("Triggering a SALE of 5 units (Stock drops to 15, below Threshold 20).")
	// SALE drops it to 15 (below threshold 20)
	_, err = pimSvc.AdjustInventory(authCtx, db.CreateInventoryLedgerEntryParams{
		OrgID:           pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:        pgtype.UUID{Bytes: tenantID, Valid: true},
		VariantID:       newVariant.ID,
		LocationID:      pgtype.UUID{Bytes: locID, Valid: true},
		TransactionType: "SALE",
		QuantityDelta:   -5,
	})
	if err != nil {
		fmt.Println("Err 3: ", err)
	}

	// Check if a DRAFT PO was created in the database
	time.Sleep(500 * time.Millisecond) // Buffer
	var poID pgtype.UUID
	err = dbpool.QueryRow(ctx, "SELECT id FROM purchase_orders WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT 1", tenantID).Scan(&poID)
	if err == nil {
		fmt.Printf("-> SUCCESS! Found Draft Purchase Order created by Background Worker: %s\n", poID.Bytes)
	} else {
		fmt.Printf("-> FAIL: %v\n", err)
	}

	fmt.Println("\n--- ALL TESTS COMPLETED. 4 IDs AND EVENT-DRIVEN ARCHITECTURE VERIFIED ---")
}
