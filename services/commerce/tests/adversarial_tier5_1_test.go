package tests

import (
	"context"
	"errors"
	"math"
	"testing"

	"commerce_modules/internal/inventory"
	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"
	"commerce_modules/internal/unified"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ = mock.Anything

func TestOMSFulfillOrderRollback_Adversarial(t *testing.T) {
	repo := new(MockRepo)
	inv := new(MockInventoryClient)
	cat := new(MockCatalogClient)
	svc := oms.NewOMSService(repo, inv, cat)

	ctx := context.Background()
	tenantID, orgID, orderID := uuid.New(), uuid.New(), uuid.New()
	items := []models.OrderLineItem{{Quantity: 1}}

	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return(items, nil)
	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusFulfilled).Return(nil)
	inv.On("DeductInventory", ctx, tenantID, orgID, orderID, items).Return(errors.New("inventory deduct error"))

	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusFulfilled, models.OrderStatusPending).Return(nil)

	err := svc.FulfillOrder(ctx, tenantID, orgID, orderID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inventory deduct error")

	repo.AssertExpectations(t)
	inv.AssertExpectations(t)
}

func TestOMSCancelOrderRollback_Adversarial(t *testing.T) {
	repo := new(MockRepo)
	inv := new(MockInventoryClient)
	cat := new(MockCatalogClient)
	svc := oms.NewOMSService(repo, inv, cat)

	ctx := context.Background()
	tenantID, orgID, orderID := uuid.New(), uuid.New(), uuid.New()
	items := []models.OrderLineItem{{Quantity: 1}}

	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return(items, nil)
	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusCancelled).Return(nil)
	inv.On("ReleaseInventory", ctx, tenantID, orgID, orderID, items).Return(errors.New("inventory release error"))

	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusCancelled, models.OrderStatusPending).Return(nil)

	err := svc.CancelOrder(ctx, tenantID, orgID, orderID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inventory release error")

	repo.AssertExpectations(t)
	inv.AssertExpectations(t)
}

func TestInventory_ReserveStock_DBError_Adversarial(t *testing.T) {
	config := embeddedpostgres.DefaultConfig().
		Port(5441).
		DataPath(t.TempDir() + "/data").
		RuntimePath(t.TempDir() + "/runtime")
	ep := embeddedpostgres.NewDatabase(config)
	if err := ep.Start(); err != nil {
		t.Fatalf("failed to start embedded db: %v", err)
	}
	defer ep.Stop()

	connStr := "postgres://postgres:postgres@localhost:5441/postgres?sslmode=disable"
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

	svc := inventory.NewPgService(pool)
	ctx := context.Background()
	pid := "overflow-test"

	err = svc.AdjustStock(ctx, uuid.Nil, uuid.Nil, pid, 100)
	assert.NoError(t, err)

	err = svc.ReserveStock(ctx, uuid.Nil, uuid.Nil, pid, math.MaxInt)
	assert.Error(t, err)
	assert.NotEqual(t, inventory.ErrInvalidQuantity, err, "database errors should not be masked as ErrInvalidQuantity")
}

type MockCustomPIMClient struct {
	mock.Mock
}

func (m *MockCustomPIMClient) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, tenantID, orgID, variantID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCustomPIMClient) GetProduct(ctx context.Context, tenantID, orgID, productID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, tenantID, orgID, productID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCustomPIMClient) ListVariants(ctx context.Context, tenantID, orgID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	args := m.Called(ctx, tenantID, orgID, productID)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestUnifiedPushProduct_Adversarial(t *testing.T) {
	pim := new(MockCustomPIMClient)
	uniClient := unified.NewMockUnifiedClient()
	uniClient.PushErr = errors.New("unified push error")

	svc := unified.NewSyncService(uniClient, pim, nil, nil)
	ctx := context.Background()
	prodID := uuid.New()

	pim.On("GetProduct", ctx, uuid.Nil, uuid.Nil, prodID).Return(nil, errors.New("not found")).Once()
	err := svc.PushProduct(ctx, uuid.Nil, uuid.Nil, "conn-1", prodID.String())
	assert.Error(t, err)
	assert.Equal(t, unified.ErrProductNotFound, err)

	pim.On("GetProduct", ctx, uuid.Nil, uuid.Nil, prodID).Return(&models.Product{}, nil).Once()
	pim.On("ListVariants", ctx, uuid.Nil, uuid.Nil, prodID).Return([]*models.ProductVariant{}, nil).Once()

	err = svc.PushProduct(ctx, uuid.Nil, uuid.Nil, "conn-1", prodID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unified push error")
}
