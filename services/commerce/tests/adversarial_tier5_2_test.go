package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"

	"commerce_modules/internal/inventory"
	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"
	"commerce_modules/internal/pim"
	"commerce_modules/internal/unified"
)

// --- OMS Mocks ---
type MockOMSRepo struct {
	oms.Repository
	GetOrderFunc func(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error)
}

func (m *MockOMSRepo) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	if m.GetOrderFunc != nil {
		return m.GetOrderFunc(ctx, tenantID, orgID, orderID)
	}
	return nil, pgx.ErrNoRows
}

// --- PIM Mocks ---
type MockPIMDB struct {
	pim.DBPool
}

type MockPIMRepo struct {
	GetProductFunc func() (*models.Product, error)
}

func (m *MockPIMRepo) GetProduct(ctx context.Context, db pim.DBTX, orgID, tenantID, productID uuid.UUID) (*models.Product, error) {
	if m.GetProductFunc != nil {
		return m.GetProductFunc()
	}
	return nil, pim.ErrNotFound
}

// --- Tests ---

// 1. OMS: GetOrder returns 500 instead of 404 on ErrNoRows
func TestAdversarial_OMS_GetOrder_ErrNoRows(t *testing.T) {
	mockRepo := &MockOMSRepo{
		GetOrderFunc: func(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
			return nil, pgx.ErrNoRows
		},
	}
	svc := oms.NewOMSService(mockRepo, nil, nil)
	api := oms.NewAPI(svc)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/oms/orders/"+uuid.New().String(), nil)
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	req.Header.Set("X-Org-ID", uuid.New().String())
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404 Not Found but got 500 Internal Server Error due to string matching bug")
}

// 2. PIM: Missing Tenant/Org IDs default to uuid.Nil and don't fail auth
func TestAdversarial_PIM_MissingTenantHeaders(t *testing.T) {
	api := pim.NewAPI(nil)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/pim/products/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected 401 Unauthorized for missing headers, but got bypass")
}

// 3. Inventory: AdjustStock creates stock for phantom product IDs by falling back to MD5
type MockInvSvc struct {
	inventory.Service
}

func (m *MockInvSvc) AdjustStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string, quantity int) error {
	return nil
}

func TestAdversarial_Inventory_AdjustStock_PhantomProduct(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"product_id": "not-a-uuid-just-a-string",
		"quantity":   100,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := inventory.NewHandler(&MockInvSvc{})
	handler.AdjustStock(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request for invalid UUID format, but handler accepted it")
}

// 4. Unified: Webhook missing signature validation or erroring out
func TestAdversarial_Unified_Webhook(t *testing.T) {
	api := unified.NewAPI(unified.NewSyncService(nil, nil, nil, nil))
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/unified/webhook", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Expected failure for empty webhook, but got OK")
}
