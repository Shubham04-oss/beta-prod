package oms

import (
	"context"
	"errors"
	"testing"

	"commerce_modules/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInventoryClient struct {
	mock.Mock
}

func (m *MockInventoryClient) ReserveInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, orderID, items)
	return args.Error(0)
}

func (m *MockInventoryClient) DeductInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, orderID, items)
	return args.Error(0)
}

func (m *MockInventoryClient) ReleaseInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, orderID, items)
	return args.Error(0)
}

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateOrderWithLineItems(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, order, items)
	return args.Error(0)
}

func (m *MockRepo) UpdateOrderStatus(ctx context.Context, tenantID, orgID, orderID uuid.UUID, currentStatus, newStatus models.OrderStatus) error {
	args := m.Called(ctx, tenantID, orgID, orderID, currentStatus, newStatus)
	return args.Error(0)
}

func (m *MockRepo) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockRepo) GetOrderLineItems(ctx context.Context, tenantID, orgID, orderID uuid.UUID) ([]models.OrderLineItem, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	return args.Get(0).([]models.OrderLineItem), args.Error(1)
}

func (m *MockRepo) CreateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	args := m.Called(ctx, tenantID, orgID, customer)
	return args.Error(0)
}

func (m *MockRepo) GetCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) (*models.Customer, error) {
	args := m.Called(ctx, tenantID, orgID, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockRepo) UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	args := m.Called(ctx, tenantID, orgID, customer)
	return args.Error(0)
}

func (m *MockRepo) DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error {
	args := m.Called(ctx, tenantID, orgID, customerID)
	return args.Error(0)
}

type MockCatalogClient struct {
	mock.Mock
}

func (m *MockCatalogClient) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, tenantID, orgID, variantID)
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

func TestOMSService_CreateOrder_EdgeCase_UnhandledUpdateError(t *testing.T) {
	repo := new(MockRepo)
	inv := new(MockInventoryClient)
	cat := new(MockCatalogClient)
	svc := NewOMSService(repo, inv, cat)

	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()
	order := &models.Order{
		TenantID: tenantID,
		OrgID:    orgID,
	}
	items := []models.OrderLineItem{}
	repo.On("CreateOrderWithLineItems", ctx, tenantID, orgID, mock.Anything, items).Return(nil)

	// Simulate ReserveInventory failure
	invErr := errors.New("inventory reservation failed")
	inv.On("ReserveInventory", ctx, tenantID, orgID, mock.Anything, items).Return(invErr)

	// Simulate UpdateOrderStatus also failing
	updateErr := errors.New("database update failed")
	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, mock.Anything, models.OrderStatusPending, models.OrderStatusFailed).Return(updateErr)

	err := svc.CreateOrder(ctx, tenantID, orgID, order, items)

	// We expect the error returned to be a combination of both errors
	assert.Contains(t, err.Error(), "inventory reservation failed", "Service should return the inventory error string")
	assert.Contains(t, err.Error(), "database update failed", "Service should return the rollback error string")

	// Is the updateErr logged or handled anywhere?
	// It is completely silently ignored in the current implementation.
	// We can verify this test passes, proving the error is swallowed.
	repo.AssertExpectations(t)
	inv.AssertExpectations(t)
	cat.AssertExpectations(t)
}
