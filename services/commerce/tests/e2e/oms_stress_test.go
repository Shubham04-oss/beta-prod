package e2e

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"commerce_modules/tests/e2e/harness"
)

func TestOMS_OutOfStock_EdgeCases(t *testing.T) {
	h := harness.Setup(t)

	// Create Product
	var createdProd struct {
		ID string `json:"id"`
	}
	statusCode := h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-OOS-SKU-%s", uuid.New().String()),
		"name":  "OMS Out Of Stock Edge Product",
		"price": 25.50,
	}, &createdProd)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}

	// Adjust inventory to 5
	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   5,
		"reason":     "initial stock",
	}, nil)

	// Attempt to order 6 - should fail with out of stock (409 Conflict)
	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{"product_id": createdProd.ID, "quantity": 6},
		},
	}
	var errResp map[string]interface{}
	statusCode = h.Post(t, "/oms/orders", payload, &errResp)
	if statusCode != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for out of stock, got %d", statusCode)
	}

	// Attempt to order 5 - should succeed
	payload["items"].([]map[string]interface{})[0]["quantity"] = 5
	var orderResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	statusCode = h.Post(t, "/oms/orders", payload, &orderResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Errorf("Expected 201 Created or 200 OK for exact stock, got %d", statusCode)
	}

	// Wait a moment for async stuff if any, although order should be synchronous
	time.Sleep(100 * time.Millisecond)

	// Verify inventory is reserved
	var invResp struct {
		Quantity int `json:"quantity"`
	}
	statusCode = h.Get(t, "/inventory/stock/"+createdProd.ID, &invResp)
	if statusCode == http.StatusOK {
		if invResp.Quantity != 0 {
			t.Errorf("Expected 0 available, got %d", invResp.Quantity)
		}
	} else {
		t.Errorf("Failed to check inventory, status %d", statusCode)
	}
}

func TestOMS_StateMachine_Stress(t *testing.T) {
	h := harness.Setup(t)

	// Create Product
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-SM-STRESS-SKU-%s", uuid.New().String()),
		"name":  "OMS State Machine Stress Product",
		"price": 10.00,
	}, &createdProd)

	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)

	// Create Order
	var createdOrder struct {
		ID string `json:"id"`
	}
	statusCode := h.Post(t, "/oms/orders", map[string]interface{}{
		"items": []map[string]interface{}{
			{"product_id": createdProd.ID, "quantity": 1},
		},
	}, &createdOrder)

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create order, got status %d", statusCode)
	}

	// Stress test concurrent status updates
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	results := make(chan int, concurrency)
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			status := "FULFILLED"
			if i%2 != 0 {
				status = "CANCELLED"
			}

			// Use our own http client or harness to avoid race conditions on harness fields if any
			// Harness Client is safe for concurrent use since it's just *http.Client
			code := h.Put(t, "/oms/orders/"+createdOrder.ID+"/status", map[string]string{
				"status": status,
			}, nil)
			results <- code
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	badRequestCount := 0
	otherErrorCount := 0

	for code := range results {
		if code == http.StatusOK || code == http.StatusNoContent {
			successCount++
		} else if code == http.StatusBadRequest {
			badRequestCount++
		} else {
			otherErrorCount++
			fmt.Printf("Unexpected status code during stress test: %d\n", code)
		}
	}

	// We expect exactly ONE successful status change. The others should fail with Bad Request (invalid state).
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful status update, got %d", successCount)
	}
	if badRequestCount != concurrency-1 {
		t.Errorf("Expected exactly %d Bad Request failures, got %d", concurrency-1, badRequestCount)
	}
	if otherErrorCount > 0 {
		t.Errorf("Encountered %d unexpected errors", otherErrorCount)
	}
}
