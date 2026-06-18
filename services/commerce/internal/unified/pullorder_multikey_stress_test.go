package unified

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPullOrder_MassiveMultiKeyConcurrency(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &AdversarialMockOMSClient{}
	inv := &MockInventoryClient{}

	svc := NewSyncService(uc, pim, oms, inv)

	keys := 10
	for i := 0; i < keys; i++ {
		orderID := fmt.Sprintf("ext-order-%d", i)
		uc.Orders[orderID] = map[string]interface{}{
			"id":          orderID,
			"status":      "PAID",
			"total_price": float64(100.5),
		}
	}

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < keys; i++ {
		orderID := fmt.Sprintf("ext-order-%d", i)
		for j := 0; j < 10; j++ { // 10 concurrent requests per key
			wg.Add(1)
			go func(oid string) {
				defer wg.Done()
				svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", oid)
			}(orderID)
		}
	}

	wg.Wait()
	duration := time.Since(start)

	if len(oms.orders) != keys {
		t.Fatalf("Expected %d orders to be created, but got %d", keys, len(oms.orders))
	}

	// Since they run concurrently across keys, duration should be ~100ms, not 10 * 100ms = 1s.
	if duration >= 500*time.Millisecond {
		t.Fatalf("Performance Bug: Multi-key concurrency failed. Duration %v is too high.", duration)
	} else {
		t.Logf("Massive Multi-Key concurrent execution successful. Duration: %v", duration)
	}
}
