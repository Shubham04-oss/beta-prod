package e2e

import (
	"commerce_modules/tests/e2e/harness"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestVerification_OMS_Cancel_RestoresInventory(t *testing.T) {
	h := harness.Setup(t)
	// 1. Create Product
	var createdProd struct {
		ID string `json:"id"`
	}
	statusCode := h.Post(t, "/pim/products", map[string]interface{}{
		"sku":   fmt.Sprintf("OMS-VER-SKU-1-%s", uuid.New().String()),
		"name":  "OMS Verification Product 1",
		"price": 10.0,
	}, &createdProd)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Fatalf("Failed to create product, got %d", statusCode)
	}

	// 2. Add Inventory
	code := h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": createdProd.ID,
		"quantity":   100,
		"reason":     "initial stock",
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("Failed to adjust inventory, got %d", code)
	}

	// 3. Check Initial Inventory
	var initialStock struct {
		Quantity int `json:"quantity"`
	}
	h.Get(t, "/inventory/stock/"+createdProd.ID, &initialStock)
	if initialStock.Quantity != 100 {
		t.Fatalf("Expected 100 initial stock, got %d", initialStock.Quantity)
	}

	// 4. Create Order
	var createdOrder struct {
		ID string `json:"id"`
	}
	statusCode = h.Post(t, "/oms/orders", Order{
		Items: []OrderItem{{ProductID: createdProd.ID, Quantity: 10}},
	}, &createdOrder)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create order, got %d", statusCode)
	}

	time.Sleep(100 * time.Millisecond) // Give db a moment just in case

	// 5. Check Reserved Inventory (Available should be 90)
	var reservedStock struct {
		Quantity int `json:"quantity"`
	}
	h.Get(t, "/inventory/stock/"+createdProd.ID, &reservedStock)
	if reservedStock.Quantity != 90 {
		t.Fatalf("Expected 90 available stock after reservation, got %d", reservedStock.Quantity)
	}

	// 6. Cancel Order
	payload := map[string]string{
		"status": "cancelled",
	}
	var resp Order
	statusCode = h.Put(t, "/oms/orders/"+createdOrder.ID+"/status", payload, &resp)
	if statusCode != http.StatusOK {
		t.Fatalf("Failed to cancel order, got %d", statusCode)
	}

	time.Sleep(100 * time.Millisecond)

	// 7. Check Inventory Again (Should be restored to 100)
	var finalStock struct {
		Quantity int `json:"quantity"`
	}
	h.Get(t, "/inventory/stock/"+createdProd.ID, &finalStock)

	if finalStock.Quantity != 100 {
		t.Fatalf("BUG IDENTIFIED: Inventory was not restored on cancellation! Expected 100, got %d", finalStock.Quantity)
	}
}
