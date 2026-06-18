package tests

import (
	"context"
	"errors"
	"testing"

	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	if args.Get(0) != nil {
		return args.Get(0).(*models.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepo) UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	args := m.Called(ctx, tenantID, orgID, customer)
	return args.Error(0)
}

func (m *MockRepo) DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error {
	args := m.Called(ctx, tenantID, orgID, customerID)
	return args.Error(0)
}

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

type MockCatalogClient struct {
	mock.Mock
}

func (m *MockCatalogClient) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, tenantID, orgID, variantID)
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

func TestOMSRollback_CreateOrder(t *testing.T) {
	repo := new(MockRepo)
	inv := new(MockInventoryClient)
	cat := new(MockCatalogClient)
	svc := oms.NewOMSService(repo, inv, cat)

	ctx := context.Background()
	order := &models.Order{
		TenantID: uuid.New(),
		OrgID:    uuid.New(),
	}
	variantID := uuid.New()
	items := []models.OrderLineItem{
		{VariantID: &variantID, Quantity: 1},
	}

	cat.On("GetVariant", ctx, order.TenantID, order.OrgID, variantID).Return(&models.ProductVariant{Price: decimal.NewFromInt(50)}, nil)
	repo.On("CreateOrderWithLineItems", ctx, order.TenantID, order.OrgID, mock.Anything, mock.Anything).Return(nil)
	inv.On("ReserveInventory", ctx, order.TenantID, order.OrgID, mock.Anything, mock.Anything).Return(errors.New("inventory unavailable"))
	repo.On("UpdateOrderStatus", ctx, order.TenantID, order.OrgID, mock.Anything, models.OrderStatusPending, models.OrderStatusFailed).Return(nil)

	err := svc.CreateOrder(ctx, order.TenantID, order.OrgID, order, items)
	assert.Error(t, err)

	repo.AssertCalled(t, "UpdateOrderStatus", ctx, order.TenantID, order.OrgID, mock.Anything, models.OrderStatusPending, models.OrderStatusFailed)
	cat.AssertExpectations(t)
}
