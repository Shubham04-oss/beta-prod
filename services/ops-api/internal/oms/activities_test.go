package oms_test

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/synq/ops-api/internal/oms"
	"github.com/synq/pkg/authcontext"
)

func TestConcurrentInventoryReservation(t *testing.T) {
	_ = godotenv.Load("../../.env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL must be set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	// 1. Setup Test Data (Tenant, Org, Location, Product, Variant, InventoryLevel)
	tenantID := uuid.New().String()
	orgID := uuid.New().String()
	locID := uuid.New().String()
	prodID := uuid.New().String()
	varID := uuid.New().String()

	// We inject the TenantID into context so oms.Repository applies RLS!
	// This proves our RLS bugfix actually works.
	ctx = context.WithValue(ctx, authcontext.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, authcontext.OrgIDKey, orgID)

	// Clean up after test
	defer func() {
		// Bypass RLS for cleanup
		pool.Exec(context.Background(), "DELETE FROM inventory_levels WHERE tenant_id = $1", tenantID)
		pool.Exec(context.Background(), "DELETE FROM product_variants WHERE id = $1", varID)
		pool.Exec(context.Background(), "DELETE FROM products WHERE id = $1", prodID)
		pool.Exec(context.Background(), "DELETE FROM locations WHERE id = $1", locID)
		pool.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
		pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", orgID)
	}()

	// Insert base data
	_, err = pool.Exec(context.Background(), "INSERT INTO organizations (id, name) VALUES ($1, 'Test Org')", orgID)
	if err == nil {
		_, err = pool.Exec(context.Background(), "INSERT INTO tenants (id, org_id, name) VALUES ($1, $2, 'Test Tenant')", tenantID, orgID)
	}
	if err == nil {
		_, err = pool.Exec(context.Background(), "INSERT INTO locations (id, tenant_id, org_id, name) VALUES ($1, $2, $3, 'Test Location')", locID, tenantID, orgID)
	}
	if err == nil {
		_, err = pool.Exec(context.Background(), "INSERT INTO products (id, tenant_id, org_id, title) VALUES ($1, $2, $3, 'Test Product')", prodID, tenantID, orgID)
	}
	if err == nil {
		_, err = pool.Exec(context.Background(), "INSERT INTO product_variants (id, tenant_id, org_id, product_id, sku, price) VALUES ($1, $2, $3, $4, 'TEST-SKU', 10.00)", varID, tenantID, orgID, prodID)
	}
	if err == nil {
		_, err = pool.Exec(context.Background(), "INSERT INTO inventory_levels (tenant_id, org_id, variant_id, location_id, available_quantity, reserved_quantity) VALUES ($1, $2, $3, $4, 100, 0)", tenantID, orgID, varID, locID)
	}
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	// 2. Initialize Activities
	repo := oms.NewRepository(pool)
	activities := oms.NewActivities(repo)

	// 3. Concurrently Reserve Inventory
	// We have 100 items. We spawn 10 workers trying to reserve 15 items each (150 total).
	// Exactly 6 workers should succeed (90 items). 4 workers should fail.
	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			reservations := []oms.LineItemReservation{
				{
					VariantID:  varID,
					LocationID: locID,
					Quantity:   15,
				},
			}

			// Simulated Order ID
			orderID := uuid.New().String()

			// Call Activity
			err := activities.ReserveInventoryActivity(ctx, orderID, reservations)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				t.Logf("Reservation failed: %v", err)
				atomic.AddInt32(&failCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// 4. Validate Constraints
	if successCount != 6 {
		t.Errorf("Expected exactly 6 successful reservations, got %d", successCount)
	}
	if failCount != 4 {
		t.Errorf("Expected exactly 4 failed reservations, got %d", failCount)
	}

	// Query final reserved quantity
	var reservedQty int
	err = pool.QueryRow(ctx, "SELECT reserved_quantity FROM inventory_levels WHERE variant_id = $1 AND location_id = $2", varID, locID).Scan(&reservedQty)
	if err != nil {
		t.Fatalf("Failed to query final reserved quantity: %v", err)
	}

	if reservedQty != 90 {
		t.Errorf("Expected exactly 90 reserved units, got %d", reservedQty)
	}
}
