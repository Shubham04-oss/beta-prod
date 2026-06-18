package oms_test

import (
	"context"
	"errors"
	"testing"

	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateOrderWithLineItems(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, order, items)
	return args.Error(0)
}

func (m *mockRepository) UpdateOrderStatus(ctx context.Context, tenantID, orgID, orderID uuid.UUID, currentStatus, newStatus models.OrderStatus) error {
	args := m.Called(ctx, tenantID, orgID, orderID, currentStatus, newStatus)
	return args.Error(0)
}

func (m *mockRepository) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetOrderLineItems(ctx context.Context, tenantID, orgID, orderID uuid.UUID) ([]models.OrderLineItem, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).([]models.OrderLineItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) CreateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	args := m.Called(ctx, tenantID, orgID, customer)
	return args.Error(0)
}

func (m *mockRepository) GetCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) (*models.Customer, error) {
	args := m.Called(ctx, tenantID, orgID, customerID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	args := m.Called(ctx, tenantID, orgID, customer)
	return args.Error(0)
}

func (m *mockRepository) DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error {
	args := m.Called(ctx, tenantID, orgID, customerID)
	return args.Error(0)
}

type mockInventoryClient struct {
	mock.Mock
}

func (m *mockInventoryClient) ReserveInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, orderID, items)
	return args.Error(0)
}

func (m *mockInventoryClient) DeductInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, orderID, items)
	return args.Error(0)
}

func (m *mockInventoryClient) ReleaseInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	args := m.Called(ctx, tenantID, orgID, orderID, items)
	return args.Error(0)
}

type mockCatalogClient struct {
	mock.Mock
}

func (m *mockCatalogClient) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, tenantID, orgID, variantID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func newUUIDPtr() *uuid.UUID {
	id := uuid.New()
	return &id
}

func TestOMSService_CreateOrder(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	order := &models.Order{
		TenantID: uuid.New(),
		OrgID:    uuid.New(),
	}
	variantID := newUUIDPtr()
	items := []models.OrderLineItem{
		{VariantID: variantID},
	}

	catalogClient.On("GetVariant", ctx, order.TenantID, order.OrgID, *variantID).Return(&models.ProductVariant{}, nil)
	repo.On("CreateOrderWithLineItems", ctx, order.TenantID, order.OrgID, order, items).Return(nil).Run(func(args mock.Arguments) {
		o := args.Get(3).(*models.Order)
		if o.Status != models.OrderStatusPending {
			t.Errorf("expected status PENDING, got %v", o.Status)
		}
		if o.ID == uuid.Nil {
			t.Errorf("expected order ID to be set")
		}
	})
	invClient.On("ReserveInventory", ctx, order.TenantID, order.OrgID, mock.AnythingOfType("uuid.UUID"), items).Return(nil)

	err := svc.CreateOrder(ctx, order.TenantID, order.OrgID, order, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo.AssertExpectations(t)
	invClient.AssertExpectations(t)
	catalogClient.AssertExpectations(t)
}

func TestOMSService_CreateOrder_InventoryFailure(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	order := &models.Order{
		TenantID: uuid.New(),
		OrgID:    uuid.New(),
	}
	items := []models.OrderLineItem{}

	expectedErr := errors.New("inventory reservation failed")

	repo.On("CreateOrderWithLineItems", ctx, order.TenantID, order.OrgID, order, items).Return(nil)
	invClient.On("ReserveInventory", ctx, order.TenantID, order.OrgID, mock.AnythingOfType("uuid.UUID"), items).Return(expectedErr)
	repo.On("UpdateOrderStatus", ctx, order.TenantID, order.OrgID, mock.AnythingOfType("uuid.UUID"), models.OrderStatusPending, models.OrderStatusFailed).Return(nil)

	err := svc.CreateOrder(ctx, order.TenantID, order.OrgID, order, items)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	repo.AssertExpectations(t)
	invClient.AssertExpectations(t)
	catalogClient.AssertExpectations(t)
}

func TestOMSService_FulfillOrder(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()
	orderID := uuid.New()

	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusFulfilled).Return(nil)
	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return([]models.OrderLineItem{}, nil)
	invClient.On("DeductInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(nil)

	err := svc.FulfillOrder(ctx, tenantID, orgID, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo.AssertExpectations(t)
	invClient.AssertExpectations(t)
	catalogClient.AssertExpectations(t)
}

func TestOMSService_CancelOrder(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()
	orderID := uuid.New()

	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusCancelled).Return(nil)
	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return([]models.OrderLineItem{}, nil)
	invClient.On("ReleaseInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(nil)

	err := svc.CancelOrder(ctx, tenantID, orgID, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo.AssertExpectations(t)
	invClient.AssertExpectations(t)
	catalogClient.AssertExpectations(t)
}

func TestOMSService_CreateCustomer(t *testing.T) {
	repo := &mockRepository{}
	svc := oms.NewOMSService(repo, nil, nil)

	ctx := context.Background()
	customer := &models.Customer{}

	repo.On("CreateCustomer", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), customer).Return(nil).Run(func(args mock.Arguments) {
		c := args.Get(3).(*models.Customer)
		if c.ID == uuid.Nil {
			t.Errorf("expected customer ID to be set")
		}
	})

	err := svc.CreateCustomer(ctx, uuid.New(), uuid.New(), customer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	repo.AssertExpectations(t)
}
