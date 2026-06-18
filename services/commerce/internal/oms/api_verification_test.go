package oms

import (
	"bytes"
	"commerce_modules/internal/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

type mockInvClientForAPI struct{ mock.Mock }

func (m *mockInvClientForAPI) ReserveInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, orderID, items).Error(0)
}
func (m *mockInvClientForAPI) DeductInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, orderID, items).Error(0)
}
func (m *mockInvClientForAPI) ReleaseInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, orderID, items).Error(0)
}

type mockRepoForAPI struct{ mock.Mock }

func (m *mockRepoForAPI) CreateOrderWithLineItems(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	return m.Called(ctx, tenantID, orgID, order, items).Error(0)
}
func (m *mockRepoForAPI) UpdateOrderStatus(ctx context.Context, tenantID, orgID, orderID uuid.UUID, currentStatus, newStatus models.OrderStatus) error {
	return m.Called(ctx, tenantID, orgID, orderID, currentStatus, newStatus).Error(0)
}
func (m *mockRepoForAPI) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRepoForAPI) GetOrderLineItems(ctx context.Context, tenantID, orgID, orderID uuid.UUID) ([]models.OrderLineItem, error) {
	args := m.Called(ctx, tenantID, orgID, orderID)
	if args.Get(0) != nil {
		return args.Get(0).([]models.OrderLineItem), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRepoForAPI) CreateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	return m.Called(ctx, tenantID, orgID, customer).Error(0)
}
func (m *mockRepoForAPI) GetCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) (*models.Customer, error) {
	args := m.Called(ctx, tenantID, orgID, customerID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRepoForAPI) UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	return m.Called(ctx, tenantID, orgID, customer).Error(0)
}
func (m *mockRepoForAPI) DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error {
	return m.Called(ctx, tenantID, orgID, customerID).Error(0)
}

type mockCatalogClientForAPI struct{ mock.Mock }

func (m *mockCatalogClientForAPI) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	args := m.Called(ctx, tenantID, orgID, variantID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.ProductVariant), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestVerification_API_CreateOrder_Prices(t *testing.T) {
	repo := &mockRepoForAPI{}
	catalogClient := &mockCatalogClientForAPI{}
	invClient := &mockInvClientForAPI{}
	svc := NewOMSService(repo, invClient, catalogClient)
	api := NewAPI(svc)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"product_id": uuid.New().String(),
				"quantity":   2,
			},
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/oms/orders", bytes.NewReader(b))
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	req.Header.Set("X-Org-ID", uuid.New().String())

	catalogClient.On("GetVariant", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(&models.ProductVariant{
		Price:        decimal.NewFromInt(42),
		OptionValues: json.RawMessage(`{"color":"red"}`),
	}, nil)

	repo.On("CreateOrderWithLineItems", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*models.Order"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		order := args.Get(3).(*models.Order)
		items := args.Get(4).([]models.OrderLineItem)

		if !order.TotalPrice.Equal(decimal.NewFromInt(84)) {
			t.Errorf("Expected total price 84, got %s", order.TotalPrice.String())
		}

		for _, item := range items {
			if !item.PriceAtPurchase.Equal(decimal.NewFromInt(42)) {
				t.Errorf("Expected line item price 42, got %s", item.PriceAtPurchase.String())
			}
			if string(item.OptionValuesAtPurchase) != `{"color":"red"}` {
				t.Errorf("Expected line item options {\"color\":\"red\"}, got %s", string(item.OptionValuesAtPurchase))
			}
		}
	})

	invClient.On("ReserveInventory", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(nil)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	repo.AssertExpectations(t)
	invClient.AssertExpectations(t)
	catalogClient.AssertExpectations(t)
}
