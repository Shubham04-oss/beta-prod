package unified

import (
	"commerce_modules/internal/models"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type MockPIMClient struct {
	variants map[uuid.UUID]*models.ProductVariant
	products map[uuid.UUID]*models.Product
	mu       sync.Mutex
}

func (m *MockPIMClient) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.variants != nil {
		if v, ok := m.variants[variantID]; ok {
			return v, nil
		}
	}
	return nil, ErrProductNotFound
}

func (m *MockPIMClient) GetProduct(ctx context.Context, tenantID, orgID, productID uuid.UUID) (*models.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.products != nil {
		if p, ok := m.products[productID]; ok {
			return p, nil
		}
	}
	return nil, ErrProductNotFound
}

func (m *MockPIMClient) ListVariants(ctx context.Context, tenantID, orgID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*models.ProductVariant
	if m.variants != nil {
		for _, v := range m.variants {
			if v.ProductID == productID {
				res = append(res, v)
			}
		}
	}
	return res, nil
}

type MockOMSClient struct {
	orders []*models.Order
	items  [][]models.OrderLineItem
	mu     sync.Mutex
}

func (m *MockOMSClient) CreateOrder(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders = append(m.orders, order)
	m.items = append(m.items, items)
	return nil
}
func (m *MockOMSClient) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, o := range m.orders {
		if o.ID == orderID {
			return o, nil
		}
	}
	return nil, nil // Return nil, nil when not found to simulate existing order not found (since OMSClient returns *models.Order, error). Wait, existing test might expect existingOrder != nil when found.
}

type MockInventoryClient struct {
	stock map[string]int
	mu    sync.Mutex
}

func (m *MockInventoryClient) AdjustStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string, quantity int) error {
	// Delay before adjusting stock to ensure all GetStock calls have finished
	time.Sleep(100 * time.Millisecond)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stock[productID] += quantity
	return nil
}

func (m *MockInventoryClient) GetStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if val, ok := m.stock[productID]; ok {
		return val, nil
	}
	return 0, errors.New("not found")
}

func TestHandleWebhook_ConcurrentInventoryUpdate_RaceCondition(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &MockOMSClient{}
	inv := &MockInventoryClient{
		stock: make(map[string]int),
	}

	inv.stock["prod-1"] = 0

	svc := NewSyncService(uc, pim, oms, inv)

	var startWg sync.WaitGroup
	var doneWg sync.WaitGroup
	startWg.Add(1)

	for i := 0; i < 50; i++ {
		doneWg.Add(1)
		go func() {
			defer doneWg.Done()
			startWg.Wait() // wait for go signal
			payload := map[string]interface{}{
				"data": map[string]interface{}{
					"product_id": "prod-1",
					"quantity":   float64(10),
				},
			}
			svc.HandleWebhook(context.Background(), "inventory.updated", payload)
		}()
	}

	startWg.Done()
	doneWg.Wait()

	finalStock := inv.stock["prod-1"]
	fmt.Printf("Final stock is: %d\n", finalStock)
	if finalStock != 10 {
		t.Fatalf("Expected final stock to be 10, but got %d due to race condition!", finalStock)
	}
}

func TestPullOrder_IdempotencyFailure(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &MockOMSClient{}
	inv := &MockInventoryClient{}

	svc := NewSyncService(uc, pim, oms, inv)

	orderID := "ext-order-123"

	// Pre-populate mock unified client
	uc.Orders[orderID] = map[string]interface{}{
		"id":          orderID,
		"status":      "PAID",
		"total_price": float64(100.5),
	}

	// Pull the same order twice
	svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", orderID)
	svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", orderID)

	// Check the number of orders created
	if len(oms.orders) != 1 {
		t.Fatalf("Expected 1 order to be created, but got %d due to lack of idempotency!", len(oms.orders))
	}
}
