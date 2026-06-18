package main

import (
	"commerce_modules/internal/inventory"
	"commerce_modules/internal/models"
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

func getFreePort(t *testing.T) uint32 {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	defer l.Close()
	return uint32(l.Addr().(*net.TCPAddr).Port)
}

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
	adapter := &inventoryAdapter{svc: svc}
	
	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()
	orderID := uuid.New()

	productA := uuid.New()
	productB := uuid.New()

	// 1. Initial stock: A=10, B=0
	svc.AdjustStock(ctx, tenantID, orgID, productA.String(), 10)
	svc.AdjustStock(ctx, tenantID, orgID, productB.String(), 0)

	// 2. Reserve A=10
	err = svc.ReserveStock(ctx, tenantID, orgID, productA.String(), 10)
	if err != nil {
		t.Fatalf("failed to reserve A: %v", err)
	}

	items := []models.OrderLineItem{
		{VariantID: productA, Quantity: 10},
		{VariantID: productB, Quantity: 10}, // Will fail
	}

	// 3. We run adapter.DeductInventory concurrently with a steal operation
	var wg sync.WaitGroup
	wg.Add(2)
	
	stolen := false
	
	go func() {
		defer wg.Done()
		adapter.DeductInventory(ctx, tenantID, orgID, orderID, items)
	}()
	
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			err := svc.ReserveStock(ctx, tenantID, orgID, productA.String(), 10)
			if err == nil {
				stolen = true
				break
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	wg.Wait()

	if !stolen {
		t.Log("Did not manage to hit the race condition (stock wasn't stolen). This can happen in tests.")
	}

	// If rollback was completely robust, DeductInventory would have either succeeded (impossible due to B)
	// or rolled back. But since it ignores ReserveStock errors, we expect that if `stolen` is true,
	// DeductInventory failed to reserve A back.
	// But actually wait! If `stolen` is true, the thief reserved the stock.
	// So available=0, reserved=10 (owned by thief).
	// But the original owner (the order) LOST its reservation!
	// It thinks the order failed, but the reservation is gone!
	// Wait, if the order failed, the reservation SHOULD be gone?
	// Ah. DeductInventory is called when we CONFIRM the order.
	// We want to reduce reserved_quantity and not touch available_quantity.
	// The order ALREADY reserved the stock.
	// DeductInventory is rolling back? No, DeductInventory should fail if B has no stock,
	// and if B fails, the order deduction fails. The order items should REMAIN reserved!
	// Because the order hasn't been cancelled!
	
	// Wait! DeductInventory is a system-level operation. If it fails midway, it should restore the state
	// exactly as it was BEFORE DeductInventory was called.
	// Before DeductInventory: A was reserved (available=0, reserved=10).
	// If DeductInventory rolls back, A should STILL be reserved.
	// But since the thief stole the stock, the thief has the reservation,
	// and DeductInventory silently failed to restore the reservation for the original order!
	// BUT WAIT, the system doesn't track WHO holds the reservation.
	// It just tracks the aggregate reserved_quantity.
	// So if the thief stole it, reserved_quantity is still 10!
	// But now we have TWO orders that think they have 10 items reserved!
	// 1. The thief who successfully called ReserveStock(A, 10).
	// 2. The original order, whose DeductInventory failed and thinks it still holds the reservation.
	// But wait, the original order's reservation was released by DeductInventory!
	// And then it failed to reserve it again!
	// So only the thief holds it.
	// But the total reserved_quantity is 10.
	// Wait, the original order was ALREADY placed. It already successfully called ReserveInventory.
	// Now it calls DeductInventory. It fails. The original order state is still "RESERVED" or it fails to transition to "DEDUCTED".
	// But its reservation is GONE. Because DeductInventory released it, and failed to re-reserve it!
	
	// So let's check total stock.
	avail, _ := svc.GetStock(ctx, tenantID, orgID, productA.String())
	// How to get reserved stock? We need a raw query.
	var res int
	pool.QueryRow(ctx, "SELECT reserved_quantity FROM inventory_levels WHERE variant_id = $1", productA).Scan(&res)
	
	t.Logf("A available: %d, reserved: %d", avail, res)
	
	if stolen {
		if avail == 0 && res == 10 {
			t.Errorf("RACE CONDITION REPRODUCED: The rollback failed silently and allowed another transaction to steal the stock! Total reserved is 10, but two transactions think they own it (the original one that failed deduction but kept its order, and the thief).")
		}
	}
}
