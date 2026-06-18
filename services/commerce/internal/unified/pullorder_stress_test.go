package unified

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestPullOrder_MassiveConcurrency(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &AdversarialMockOMSClient{}
	inv := &MockInventoryClient{}

	svc := NewSyncService(uc, pim, oms, inv)

	orderID := "ext-order-massive"

	uc.Orders[orderID] = map[string]interface{}{
		"id":          orderID,
		"status":      "PAID",
		"total_price": float64(100.5),
	}

	var wg sync.WaitGroup
	workers := 1000 // High concurrency

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", orderID)
		}()
	}

	wg.Wait()

	if len(oms.orders) > 1 {
		t.Fatalf("Expected 1 order to be created, but got %d! Concurrent PullOrder has a TOCTOU race condition.", len(oms.orders))
	} else if len(oms.orders) == 1 {
		t.Log("Order creation was properly synchronized.")
	} else {
		t.Fatalf("No orders were created.")
	}
}
