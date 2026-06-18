package e2e

import (
	"commerce_modules/tests/e2e/harness"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestOMS_UpdateOrderStatus_ArbitraryStatus(t *testing.T) {
	h := harness.Setup(t)

	// Create Product
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-SKU-EDGE-1-%s", uuid.New().String()),
		"name":  "OMS Product Edge 1",
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
	h.Post(t, "/oms/orders", Order{
		Items: []OrderItem{{ProductID: createdProd.ID, Quantity: 2}},
	}, &createdOrder)

	// Update to an arbitrary garbage status
	payload := map[string]string{
		"status": "GARBAGE_STATUS",
	}

	var resp Order
	statusCode := h.Put(t, "/oms/orders/"+createdOrder.ID+"/status", payload, &resp)

	// It should fail because the status is invalid
	if statusCode != http.StatusBadRequest {
		t.Fatalf("Expected status Bad Request for garbage status, got %d", statusCode)
	}
}

func TestOMS_CreateOrder_WithCustomer(t *testing.T) {
	h := harness.Setup(t)
	// Create Product
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-SKU-EDGE-2-%s", uuid.New().String()),
		"name":  "OMS Product Edge 2",
		"price": 10.0,
	}, &createdProd)

	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)

	// Ensure customer_id is correctly parsed and assigned
	payload := map[string]interface{}{
		"customer_id": "00000000-0000-0000-0000-000000000000",
		"items": []map[string]interface{}{
			{"product_id": createdProd.ID, "quantity": 1},
		},
	}
	var createdOrder struct {
		ID         string  `json:"id"`
		CustomerID *string `json:"customer_id"`
	}

	statusCode := h.Post(t, "/oms/orders", payload, &createdOrder)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Expected status Created or OK, got %d", statusCode)
	}

	// The CustomerID should match the provided ID
	if createdOrder.CustomerID == nil {
		t.Errorf("Expected CustomerID to not be nil, but got nil")
	} else if *createdOrder.CustomerID != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("Expected CustomerID to be 00000000-0000-0000-0000-000000000000, got %s", *createdOrder.CustomerID)
	}
}
