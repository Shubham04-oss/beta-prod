package inventory

import (
	"context"
	"math"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPgService_Overflow(t *testing.T) {
	config := embeddedpostgres.DefaultConfig().
		Port(5433).
		DataPath(t.TempDir() + "/data").
		RuntimePath(t.TempDir() + "/runtime")
	ep := embeddedpostgres.NewDatabase(config)
	if err := ep.Start(); err != nil {
		t.Fatalf("failed to start embedded db: %v", err)
	}
	defer ep.Stop()

	connStr := "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
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

	svc := NewPgService(pool)
	ctx := context.Background()
	pid := "overflow-test"

	err = svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, pid, math.MaxInt)
	t.Logf("AdjustStock math.MaxInt error: %v", err)

	err = svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, pid, 2000000000)
	t.Logf("AdjustStock 2B error: %v", err)

	err = svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, pid, 2000000000)
	t.Logf("AdjustStock 2B + 2B error: %v", err)

	avail, _ := svc.GetStock(ctx, uuid.Nil, uuid.Nil, pid)
	t.Logf("Available stock: %d", avail)
}
