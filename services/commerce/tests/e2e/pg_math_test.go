package e2e

import (
	"commerce_modules/internal/inventory"
	"context"
	"math"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIntegerOverflow_MathInt32(t *testing.T) {
	config := embeddedpostgres.DefaultConfig().
		Port(5434).
		DataPath(t.TempDir() + "/data").
		RuntimePath(t.TempDir() + "/runtime")
	ep := embeddedpostgres.NewDatabase(config)
	if err := ep.Start(); err != nil {
		t.Fatalf("failed to start embedded postgres: %v", err)
	}
	defer ep.Stop()

	connStr := "postgres://postgres:postgres@localhost:5434/postgres?sslmode=disable"
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
	productID := "prod-math32"

	svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, productID, 2000000000)

	// test if it overflows to negative using another big addition
	err = svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, productID, 2000000000)
	t.Logf("Result of 2B + 2B adjust: %v", err)

	stock, _ := svc.GetStock(ctx, uuid.Nil, uuid.Nil, productID)
	t.Logf("Stock after 2B+2B adjust: %d", stock)
	if stock < 0 {
		t.Errorf("FAIL: Stock wrapped to negative! %d", stock)
	}

	// Reset stock
	pool.Exec(ctx, "UPDATE inventory_levels SET available_quantity = 100")

	// Test underflow bypassing the `>= 0` check
	// If available_quantity + $1 wraps around, it might be positive, bypassing the check?
	// e.g. 100 + X = 5. So X = -95. If X wraps around...
	// Postgres uses 32-bit signed integers for INTEGER.
	err = svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, productID, math.MinInt32)
	t.Logf("Result of MinInt32 adjust: %v", err)

	stock, _ = svc.GetStock(ctx, uuid.Nil, uuid.Nil, productID)
	t.Logf("Stock after MinInt32: %d", stock)
	if stock < 0 {
		t.Errorf("FAIL: Stock negative: %d", stock)
	}
}
