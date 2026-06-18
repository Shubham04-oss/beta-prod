package e2e

import (
	"commerce_modules/internal/inventory"
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDeductRollbackRaceCondition(t *testing.T) {
	port := getFreePort(t)
	config := embeddedpostgres.DefaultConfig().
		Port(port).
		DataPath(t.TempDir() + "/data").
		RuntimePath(t.TempDir() + "/runtime")
	ep := embeddedpostgres.NewDatabase(config)
	if err := ep.Start(); err != nil {
		t.Fatalf("failed to start embedded postgres: %v", err)
	}
	defer ep.Stop()

	connStr := fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("failed to connect to embedded db: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS locations (
			id UUID PRIMARY KEY,
			org_id UUID NOT NULL,
			tenant_id UUID NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		);
		CREATE TABLE IF NOT EXISTS inventory_levels (
			id UUID PRIMARY KEY,
			org_id UUID NOT NULL,
			tenant_id UUID NOT NULL,
			variant_id UUID NOT NULL,
			location_id UUID NOT NULL,
			available_quantity INTEGER NOT NULL DEFAULT 0,
			reserved_quantity INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_by UUID,
			UNIQUE(org_id, tenant_id, location_id, variant_id)
		);
	`)
	if err != nil {
		t.Fatalf("failed to execute DDL: %v", err)
	}

	svc := inventory.NewPgService(pool)
	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()

	productA := "prod-a"
	productB := "prod-b"

	// 1. Initial stock: A=10, B=0
	svc.AdjustStock(ctx, tenantID, orgID, productA, 10)
	svc.AdjustStock(ctx, tenantID, orgID, productB, 0) // ensuring record exists

	// 2. Reserve A=10
	err = svc.ReserveStock(ctx, tenantID, orgID, productA, 10)
	if err != nil {
		t.Fatalf("failed to reserve A: %v", err)
	}

	// 3. Now we have an order with A=10, B=10
	// We want to deduct it. Since A has reserved 10, it releases A (available+10, reserved-10).
	// Then it adjusts A by -10. So A's available=0, reserved=0.
	// Then it processes B. B has 0 reserved. ReleaseStock(B) FAILS.
	// So it rolls back A.
	// Rollback A: AdjustStock(A, 10) -> available=10.
	// ReserveStock(A, 10) -> available=0, reserved=10.

	// If during the window after AdjustStock(A, 10) and before ReserveStock(A, 10),
	// another thread reserves A=10, then the ReserveStock(A, 10) in rollback will fail!
	// And since the error is ignored, we end up with available=0, reserved=0 (instead of 10) for A!

	// Wait, we need to mock or intercept to time it perfectly, or use a lot of goroutines.
	// Let's use a lot of goroutines to hit the race condition.

	// Since we don't have the adapter here, let's copy its DeductInventory logic.
	deductInventory := func() {
		// Simulate loop for A
		if err := svc.ReleaseStock(ctx, tenantID, orgID, productA, 10); err != nil {
			return
		}
		if err := svc.AdjustStock(ctx, tenantID, orgID, productA, -10); err != nil {
			return
		}

		// Simulate loop for B (fails)
		if err := svc.ReleaseStock(ctx, tenantID, orgID, productB, 10); err != nil {
			// Rollback A
			svc.AdjustStock(ctx, tenantID, orgID, productA, 10)
			
			// We sleep slightly to widen the race window
			time.Sleep(10 * time.Millisecond)

			svc.ReserveStock(ctx, tenantID, orgID, productA, 10)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		deductInventory()
	}()
	
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		// Malicious concurrent transaction steals the stock!
		svc.ReserveStock(ctx, tenantID, orgID, productA, 10)
	}()

	wg.Wait()

	// What is the state of A?
	var avail, res int
	pool.QueryRow(ctx, "SELECT available_quantity, reserved_quantity FROM inventory_levels WHERE variant_id = $1", inventory.GetVariantIDForTest(productA)).Scan(&avail, &res)

	t.Logf("Final State -> A available: %d, reserved: %d", avail, res)
	
	// Total stock should be 10.
	if avail + res != 10 {
		t.Fatalf("Data loss! Total stock for A should be 10, got %d", avail+res)
	}
}
