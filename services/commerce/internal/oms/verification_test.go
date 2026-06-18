package oms

import (
	"context"
	"testing"

	"commerce_modules/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockInvClientVerification struct{ mock.Mock }

func (m *mockInvClientVerification) ReserveInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, orderID, items).Error(0)
}
func (m *mockInvClientVerification) DeductInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, orderID, items).Error(0)
}
func (m *mockInvClientVerification) ReleaseInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, orderID, items).Error(0)
}

type mockRepoVerification struct{ mock.Mock }

func (m *mockRepoVerification) CreateOrderWithLineItems(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, order, items).Error(0)
}
func (m *mockRepoVerification) UpdateOrderStatus(ctx context.Context, tenantID, orgID, orderID uuid.UUID, currentStatus, newStatus models.OrderStatus) error {
	return m.Called(ctx, tenantID, orgID, orderID, currentStatus, newStatus).Error(0)
}
func (m *mockRepoVerification) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRepoVerification) GetOrderLineItems(ctx context.Context, tenantID, orgID, orderID uuid.UUID) ([]models.OrderLineItem, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).([]models.OrderLineItem), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRepoVerification) CreateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	return m.Called(ctx, tenantID, orgID, customer).Error(0)
}
func (m *mockRepoVerification) GetCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) (*models.Customer, error) {
	args := m.Called(ctx, tenantID, orgID, customerID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRepoVerification) UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	return m.Called(ctx, tenantID, orgID, customer).Error(0)
}
func (m *mockRepoVerification) DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error {
	return m.Called(ctx, tenantID, orgID, customerID).Error(0)
}

type mockCatalogClientVerification struct{ mock.Mock }

func (m *mockCatalogClientVerification) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, tenantID, orgID, variantID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestVerification_OMSService_CancelOrder(t *testing.T) {
	ctx := context.Background()
	tenantID, orgID := uuid.New(), uuid.New()
	orderID := uuid.New()

	repo := &mockRepoVerification{}
	invClient := &mockInvClientVerification{}
	catalogClient := &mockCatalogClientVerification{}
	svc := NewOMSService(repo, invClient, catalogClient)

	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return([]models.OrderLineItem{}, nil)
	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusCancelled).Return(nil)
	invClient.On("ReleaseInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(nil)

	err := svc.CancelOrder(ctx, tenantID, orgID, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo.AssertExpectations(t)
	invClient.AssertExpectations(t)
	catalogClient.AssertExpectations(t)
}
