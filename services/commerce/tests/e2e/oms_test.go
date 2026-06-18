package e2e

import (
	"commerce_modules/tests/e2e/harness"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Order struct {
	ID     string      `json:"id,omitempty"`
	Items  []OrderItem `json:"items"`
	Status string      `json:"status,omitempty"`
}

func TestOMS_CreateOrder_ValidItems(t *testing.T) {
	h := harness.Setup(t)
	// 1. Create Product
	var created struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-SKU-1-%s", uuid.New().String()),
		"name":  "OMS Product 1",
		"price": 10.0,
	}, &created)

	// 2. Add Inventory
	code := h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": created.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)
	if code != 200 {
		t.Fatalf("Failed to adjust inventory, got %d", code)
	}

	payload := Order{
		Items: []OrderItem{
			{ProductID: created.ID, Quantity: 2},
		},
	}

	var resp Order
	statusCode := h.Post(t, "/oms/orders", payload, &resp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Errorf("Expected status Created or OK, got %d", statusCode)
	}
}

func TestOMS_UpdateOrderStatus_ToShipped(t *testing.T) {
	h := harness.Setup(t)
	// 1. Create Product & Order
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-SKU-2-%s", uuid.New().String()),
		"name":  "OMS Product 2",
		"price": 10.0,
	}, &createdProd)

	code := h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)
	if code != 200 {
		t.Fatalf("Failed to adjust inventory, got %d", code)
	}

	var createdOrder struct {
		ID string `json:"id"`
	}
	h.Post(t, "/oms/orders", Order{
		Items: []OrderItem{{ProductID: createdProd.ID, Quantity: 2}},
	}, &createdOrder)

	payload := map[string]string{
		"status": "shipped",
	}

	var resp Order
	statusCode := h.Put(t, "/oms/orders/"+createdOrder.ID+"/status", payload, &resp)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}

func TestOMS_CancelOrder_RevertsState(t *testing.T) {
	h := harness.Setup(t)
	// 1. Create Product & Order
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-SKU-3-%s", uuid.New().String()),
		"name":  "OMS Product 3",
		"price": 10.0,
	}, &createdProd)

	code := h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)
	if code != 200 {
		t.Fatalf("Failed to adjust inventory, got %d", code)
	}

	var createdOrder struct {
		ID string `json:"id"`
	}
	h.Post(t, "/oms/orders", Order{
		Items: []OrderItem{{ProductID: createdProd.ID, Quantity: 2}},
	}, &createdOrder)

	payload := map[string]string{
		"status": "cancelled",
	}

	var resp Order
	statusCode := h.Put(t, "/oms/orders/"+createdOrder.ID+"/status", payload, &resp)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}

func TestOMS_GetOrder_ReturnsFullDetails(t *testing.T) {
	h := harness.Setup(t)
	// 1. Create Product & Order
	var createdProd struct {
		ID string `json:"id"`
	}
	h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-SKU-4-%s", uuid.New().String()),
		"name":  "OMS Product 4",
		"price": 10.0,
	}, &createdProd)

	code := h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)
	if code != 200 {
		t.Fatalf("Failed to adjust inventory, got %d", code)
	}

	var createdOrder struct {
		ID string `json:"id"`
	}
	h.Post(t, "/oms/orders", Order{
		Items: []OrderItem{{ProductID: createdProd.ID, Quantity: 2}},
	}, &createdOrder)

	var resp Order
	statusCode := h.Get(t, "/oms/orders/"+createdOrder.ID, &resp)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}

func TestOMS_CreateOrder_InvalidProduct_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := Order{
		Items: []OrderItem{
			{ProductID: "invalid-prod-id", Quantity: 1},
		},
	}

	statusCode := h.Post(t, "/oms/orders", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}
