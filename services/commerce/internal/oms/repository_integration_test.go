package oms_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

func getFreePort() uint32 {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return uint32(l.Addr().(*net.TCPAddr).Port)
}

func setupDB(t *testing.T) (*embeddedpostgres.EmbeddedPostgres, *pgxpool.Pool) {
	port := getFreePort()
	config := embeddedpostgres.DefaultConfig().
		Port(port).
		DataPath(t.TempDir() + "/data").
		RuntimePath(t.TempDir() + "/runtime")

	db := embeddedpostgres.NewDatabase(config)
	err := db.Start()
	if err != nil {
		t.Fatalf("Failed to start embedded postgres: %v", err)
	}

	ctx := context.Background()
	connStr := fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres", port)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		db.Stop()
		t.Fatalf("Failed to connect to pool: %v", err)
	}

	schema := `
	CREATE TABLE orders (
		id UUID PRIMARY KEY,
		org_id UUID NOT NULL,
		tenant_id UUID NOT NULL,
		customer_id UUID,
		currency VARCHAR(10),
		status VARCHAR(20),
		total_price NUMERIC,
		metadata JSONB,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	);

	CREATE TABLE order_line_items (
		id UUID PRIMARY KEY,
		org_id UUID NOT NULL,
		tenant_id UUID NOT NULL,
		order_id UUID NOT NULL,
		variant_id UUID,
		price_at_purchase NUMERIC,
		option_values_at_purchase JSONB,
		quantity INT,
		created_at TIMESTAMP
	);

	CREATE TABLE customers (
		id UUID PRIMARY KEY,
		org_id UUID NOT NULL,
		tenant_id UUID NOT NULL,
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		email VARCHAR(255),
		metadata JSONB,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	);
	`
	_, err = pool.Exec(ctx, schema)
	if err != nil {
		pool.Close()
		db.Stop()
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db, pool
}

func ptrStr(s string) *string        { return &s }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

func TestPostgresRepository_RLS_Enforcement_And_Idempotency(t *testing.T) {
	db, pool := setupDB(t)
	defer db.Stop()
	defer pool.Close()

	repo := oms.NewPostgresRepository(pool)
	ctx := context.Background()

	realTenantID := uuid.New()
	realOrgID := uuid.New()

	fakeTenantID := uuid.New()
	fakeOrgID := uuid.New()

	cust := &models.Customer{
		ID:        uuid.New(),
		TenantID:  fakeTenantID,
		OrgID:     fakeOrgID,
		FirstName: ptrStr("John"),
	}

	err := repo.CreateCustomer(ctx, realTenantID, realOrgID, cust)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	savedCust, err := repo.GetCustomer(ctx, realTenantID, realOrgID, cust.ID)
	if err != nil {
		t.Fatalf("Failed to get customer: %v", err)
	}

	if savedCust.TenantID != realTenantID || savedCust.OrgID != realOrgID {
		t.Errorf("RLS bypass detected in CreateCustomer! Expected TenantID %v, got %v", realTenantID, savedCust.TenantID)
	}

	order := &models.Order{
		ID:         uuid.New(),
		TenantID:   fakeTenantID,
		OrgID:      fakeOrgID,
		CustomerID: ptrUUID(cust.ID),
		Status:     models.OrderStatusPending,
		TotalPrice: decimal.NewFromInt(100),
	}

	item := models.OrderLineItem{
		ID:        uuid.New(),
		TenantID:  fakeTenantID,
		OrgID:     fakeOrgID,
		VariantID: ptrUUID(uuid.New()),
		Quantity:  1,
	}

	err = repo.CreateOrderWithLineItems(ctx, realTenantID, realOrgID, order, []models.OrderLineItem{item})
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	savedOrder, err := repo.GetOrder(ctx, realTenantID, realOrgID, order.ID)
	if err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}

	if savedOrder.TenantID != realTenantID || savedOrder.OrgID != realOrgID {
		t.Errorf("RLS bypass detected in CreateOrder! Expected TenantID %v, got %v", realTenantID, savedOrder.TenantID)
	}

	// Now test Idempotency
	err = repo.UpdateOrderStatus(ctx, realTenantID, realOrgID, order.ID, models.OrderStatusPending, models.OrderStatusFulfilled)
	if err != nil {
		t.Fatalf("Expected valid update to succeed, got %v", err)
	}

	err = repo.UpdateOrderStatus(ctx, realTenantID, realOrgID, order.ID, models.OrderStatusPending, models.OrderStatusCancelled)
	if err == nil {
		t.Fatalf("Expected error due to invalid state transition (idempotency), but update succeeded")
	}
}
