package e2e

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/google/uuid"

	"commerce_modules/tests/e2e/harness"
)

func TestOMS_Oracle_GarbageStatusUpdates(t *testing.T) {
	h := harness.Setup(t)

	// Generator for products
	sku := fmt.Sprintf("STRESS-SKU-%s", uuid.New().String())
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   sku,
		"name":  "Stress Product",
		"price": 10.0,
	}, &createdProd)

	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)

	var createdOrder struct {
		ID string `json:"id"`
	}
	statusCode := h.Post(t, "/oms/orders", Order{
		Items: []OrderItem{{ProductID: createdProd.ID, Quantity: 2}},
	}, &createdOrder)

	if statusCode != http.StatusCreated {
		t.Fatalf("Failed to create order, got %d", statusCode)
	}

	// Generator for garbage statuses
	garbageStatuses := []string{
		"GARBAGE",
		"PAID_BUT_NOT_REALLY",
		"PENDINGG",
		"",
		"   ",
		"DROP TABLE orders;",
		"{}",
	}

	for _, status := range garbageStatuses {
		payload := map[string]string{
			"status": status,
		}
		var resp Order
		code := h.Put(t, "/oms/orders/"+createdOrder.ID+"/status", payload, &resp)

		// Oracle: System must gracefully reject garbage statuses with 400 Bad Request
		if code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for garbage status '%s', got %d", status, code)
		}
	}
}

func TestOMS_Oracle_OutOfStockOrders(t *testing.T) {
	h := harness.Setup(t)

	sku := fmt.Sprintf("OOS-SKU-%s", uuid.New().String())
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   sku,
		"name":  "Out of Stock Product",
		"price": 15.0,
	}, &createdProd)

	// Set initial inventory to 5
	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   5,
		"reason":     "initial stock",
	}, nil)

	// Generator for valid and out-of-stock order quantities
	tests := []struct {
		qty            int
		expectSuccess  bool
		expectedStatus int
	}{
		{qty: 2, expectSuccess: true, expectedStatus: http.StatusCreated},   // Valid, 3 left
		{qty: 4, expectSuccess: false, expectedStatus: http.StatusConflict}, // Invalid, requests 4 but only 3 left
		{qty: 3, expectSuccess: true, expectedStatus: http.StatusCreated},   // Valid, exactly 3 left (0 left)
		{qty: 1, expectSuccess: false, expectedStatus: http.StatusConflict}, // Invalid, 0 left
	}

	for i, tc := range tests {
		var createdOrder struct {
			ID string `json:"id"`
		}
		code := h.Post(t, "/oms/orders", Order{
			Items: []OrderItem{{ProductID: createdProd.ID, Quantity: tc.qty}},
		}, &createdOrder)

		if tc.expectSuccess && code != tc.expectedStatus {
			t.Errorf("Test %d: expected success (%d), got %d", i, tc.expectedStatus, code)
		} else if !tc.expectSuccess && code != tc.expectedStatus {
			// Oracle: OOS should give 409 Conflict (or 400 Bad Request based on implementation, let's accept either)
			if code != http.StatusConflict && code != http.StatusBadRequest {
				t.Errorf("Test %d: expected 409 or 400 for out of stock, got %d", i, code)
			}
		}
	}
}

func TestOMS_Stress_ConcurrentStateTransitions(t *testing.T) {
	h := harness.Setup(t)

	sku := fmt.Sprintf("STRESS-CONC-%s", uuid.New().String())
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   sku,
		"name":  "Concurrent Product",
		"price": 50.0,
	}, &createdProd)

	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   1000,
		"reason":     "large stock",
	}, nil)

	var createdOrder struct {
		ID string `json:"id"`
	}
	code := h.Post(t, "/oms/orders", Order{
		Items: []OrderItem{{ProductID: createdProd.ID, Quantity: 10}},
	}, &createdOrder)

	if code != http.StatusCreated {
		t.Fatalf("Failed to create order, got %d", code)
	}

	// Generator: concurrently fire valid and invalid state transitions
	var wg sync.WaitGroup
	workers := 20
	results := make([]int, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Some workers will try to FULFILL, others CANCEL
			status := "FULFILLED"
			if idx%2 == 0 {
				status = "CANCELLED"
			}

			payload := map[string]string{
				"status": status,
			}
			var resp Order
			resCode := h.Put(t, "/oms/orders/"+createdOrder.ID+"/status", payload, &resp)
			results[idx] = resCode
		}(i)
	}

	wg.Wait()

	// Oracle:
	// Exactly ONE transition should succeed (200 OK)
	// The rest should fail with 400 Bad Request (invalid state transition)
	successes := 0
	failures := 0
	others := 0

	for _, resCode := range results {
		if resCode == http.StatusOK {
			successes++
		} else if resCode == http.StatusBadRequest {
			failures++
		} else {
			others++
			t.Errorf("Unexpected status code during concurrent update: %d", resCode)
		}
	}

	if successes != 1 {
		t.Errorf("Expected exactly 1 successful state transition, got %d", successes)
	}
	if failures != workers-1 {
		t.Errorf("Expected exactly %d failed state transitions, got %d", workers-1, failures)
	}
}
