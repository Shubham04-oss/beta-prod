package unified

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type MockInventoryClientWithDelay struct {
	stock map[string]int
	mu    sync.Mutex
}

func (m *MockInventoryClientWithDelay) AdjustStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stock[productID] += quantity
	return nil
}

func (m *MockInventoryClientWithDelay) GetStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string) (int, error) {
	m.mu.Lock()
	val, ok := m.stock[productID]
	m.mu.Unlock()

	// Simulate DB/Network delay
	time.Sleep(10 * time.Millisecond)

	if ok {
		return val, nil
	}
	return 0, errors.New("not found")
}

func TestHandleWebhook_RaceCondition(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &MockOMSClient{}
	inv := &MockInventoryClientWithDelay{
		stock: make(map[string]int),
	}

	svc := NewSyncService(uc, pim, oms, inv)

	productID := "prod-race"
	inv.stock[productID] = 0

	var wg sync.WaitGroup
	workers := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			payload := map[string]interface{}{
				"data": map[string]interface{}{
					"product_id": productID,
					"quantity":   float64(10),
				},
			}
			err := svc.HandleWebhook(context.Background(), "inventory.updated", payload)
			if err != nil {
				t.Errorf("Webhook error: %v", err)
			}
		}(i)
	}

	wg.Wait()

	finalStock := inv.stock[productID]
	if finalStock != 10 {
		t.Fatalf("Expected final stock to be 10, got %d. This indicates a race condition where increments compound!", finalStock)
	}
}
