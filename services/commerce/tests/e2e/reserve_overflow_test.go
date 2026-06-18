package e2e

import (
	"commerce_modules/internal/inventory"
	"context"
	"fmt"
	"net"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getFreePort(t *testing.T) uint32 {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	defer l.Close()
	return uint32(l.Addr().(*net.TCPAddr).Port)
}

func TestReserveOverflow(t *testing.T) {
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
	productID := "prod-reserve"

	// 1. Add lots of stock
	svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, productID, 2000000000)

	// 2. Reserve lots of stock
	err = svc.ReserveStock(ctx, uuid.Nil, uuid.Nil, productID, 2000000000)
	if err != nil {
		t.Fatalf("Failed to reserve: %v", err)
	}

	// 3. Add more stock
	svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, productID, 2000000000)

	// 4. Reserve more stock, causing reserved_quantity + 2B to overflow 32-bit int
	err = svc.ReserveStock(ctx, uuid.Nil, uuid.Nil, productID, 2000000000)
	t.Logf("Result of reserving again: %v", err)

	if err == nil {
		t.Errorf("FAIL: Expected error when reserved_quantity overflows!")
	}

	// Check available quantity didn't decrease
	stock, _ := svc.GetStock(ctx, uuid.Nil, uuid.Nil, productID)
	if stock != 2000000000 {
		t.Errorf("FAIL: Available stock changed! Expected 2000000000, got %d", stock)
	}
}
