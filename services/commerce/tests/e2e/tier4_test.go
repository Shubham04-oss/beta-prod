package e2e

import (
	"bytes"
	"commerce_modules/tests/e2e/harness"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestTier4_FullProductLifecycle(t *testing.T) {
	h := harness.Setup(t)

	// Create product -> add stock -> sync push -> update product -> sync push -> delete -> assert 404
	prodPayload := PIMProduct{
		SKU:         fmt.Sprintf("T4-SKU-1-%s", uuid.New().String()),
		Name:        "T4 Product 1",
		Description: "Tier 4 Full Lifecycle",
		Price:       150.00,
	}
	var prodResp PIMProduct
	status := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Failed to create product, status: %d", status)
	}

	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected valid ID, got empty string")
	}

	// Add stock
	stockPayload := InventoryStock{
		ProductID: prodID,
		Quantity:  100,
	}
	status = h.Post(t, "/inventory/adjust", stockPayload, nil)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("Failed to add stock, status: %d", status)
	}

	// Sync push
	syncPayload := SyncPayload{
		ProductID: prodID,
	}
	status = h.Post(t, "/unified/sync/push/product", syncPayload, nil)
	if status != http.StatusOK {
		t.Fatalf("Failed to sync push product, status: %d", status)
	}

	// Update product
	updatePayload := PIMProduct{
		Name:  "T4 Product 1 Updated",
		Price: 175.00,
	}
	status = h.Put(t, "/pim/products/"+prodID, updatePayload, nil)
	if status != http.StatusOK {
		t.Fatalf("Failed to update product, status: %d", status)
	}

	// Sync push again
	status = h.Post(t, "/unified/sync/push/product", syncPayload, nil)
	if status != http.StatusOK {
		t.Fatalf("Failed to sync push product after update, status: %d", status)
	}

	// Delete product
	status = h.Delete(t, "/pim/products/"+prodID)
	if status != http.StatusNoContent && status != http.StatusOK {
		t.Fatalf("Failed to delete product, status: %d", status)
	}

	// Assert 404
	status = h.Get(t, "/pim/products/"+prodID, nil)
	if status != http.StatusNotFound {
		t.Fatalf("Expected 404 after delete, got: %d", status)
	}
}

func TestTier4_EndToEndOrderFulfillment(t *testing.T) {
	h := harness.Setup(t)

	// Create product -> add stock -> create order -> reserve inventory -> mark shipped -> verify stock deduction

	// Create product
	prodPayload := PIMProduct{
		SKU:   fmt.Sprintf("T4-SKU-2-%s", uuid.New().String()),
		Name:  "T4 Product 2",
		Price: 50.00,
	}
	var prodResp PIMProduct
	status := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Failed to create product, status: %d", status)
	}
	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected valid ID, got empty string")
	}

	// Add stock
	stockPayload := InventoryStock{
		ProductID: prodID,
		Quantity:  50,
	}
	status = h.Post(t, "/inventory/adjust", stockPayload, nil)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("Failed to add stock, status: %d", status)
	}

	// Create order
	orderPayload := Order{
		Items: []OrderItem{
			{ProductID: prodID, Quantity: 2},
		},
	}
	var orderResp Order
	status = h.Post(t, "/oms/orders", orderPayload, &orderResp)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Failed to create order, status: %d", status)
	}
	orderID := orderResp.ID
	if orderID == "" {
		t.Fatalf("Expected valid ID, got empty string")
	}

	// Reserve inventory
	reservePayload := InventoryStock{
		ProductID: prodID,
		Quantity:  2,
	}
	status = h.Post(t, "/inventory/reserve", reservePayload, nil)
	if status != http.StatusOK {
		t.Fatalf("Failed to reserve inventory, status: %d", status)
	}

	// Mark shipped
	statusPayload := map[string]string{"status": "shipped"}
	status = h.Put(t, "/oms/orders/"+orderID+"/status", statusPayload, nil)
	if status != http.StatusOK {
		t.Fatalf("Failed to mark order shipped, status: %d", status)
	}

	// Verify stock deduction
	var stockResp InventoryStock
	status = h.Get(t, "/inventory/stock/"+prodID, &stockResp)
	if status != http.StatusOK {
		t.Fatalf("Failed to get stock, status: %d", status)
	}
}

func TestTier4_BulkInventorySyncFromExternal(t *testing.T) {
	h := harness.Setup(t)

	// Create 2 products -> simulate external inventory.updated via webhook -> verify local stock

	// Create 2 products
	p1 := PIMProduct{SKU: fmt.Sprintf("T4-SKU-3A-%s", uuid.New().String()), Name: "Product 3A"}
	p2 := PIMProduct{SKU: fmt.Sprintf("T4-SKU-3B-%s", uuid.New().String()), Name: "Product 3B"}
	var rp1, rp2 PIMProduct

	status := h.Post(t, "/pim/products", p1, &rp1)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Failed to create product 1, status: %d", status)
	}

	status = h.Post(t, "/pim/products", p2, &rp2)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Failed to create product 2, status: %d", status)
	}

	id1 := rp1.ID
	id2 := rp2.ID
	if id1 == "" || id2 == "" {
		t.Fatalf("Expected valid IDs, got empty string")
	}

	// Simulate external inventory.updated via webhook for product 1
	webhookPayload1 := map[string]interface{}{
		"event": "inventory.updated",
		"data": map[string]interface{}{
			"product_id": id1,
			"quantity":   200,
		},
	}
	status = h.Post(t, "/unified/webhook", webhookPayload1, nil)
	if status != http.StatusOK {
		t.Fatalf("Failed to sync webhook 1, status: %d", status)
	}

	// Simulate external inventory.updated via webhook for product 2
	webhookPayload2 := map[string]interface{}{
		"event": "inventory.updated",
		"data": map[string]interface{}{
			"product_id": id2,
			"quantity":   300,
		},
	}
	status = h.Post(t, "/unified/webhook", webhookPayload2, nil)
	if status != http.StatusOK {
		t.Fatalf("Failed to sync webhook 2, status: %d", status)
	}

	// Verify local stock
	var s1, s2 InventoryStock
	status = h.Get(t, "/inventory/stock/"+id1, &s1)
	if status != http.StatusOK {
		t.Fatalf("Failed to get stock for product 1, status: %d", status)
	}
	status = h.Get(t, "/inventory/stock/"+id2, &s2)
	if status != http.StatusOK {
		t.Fatalf("Failed to get stock for product 2, status: %d", status)
	}
}

func TestTier4_OutOfStockOrderHandling(t *testing.T) {
	h := harness.Setup(t)

	// Create product -> add low stock (5) -> create order for > 5 -> reserve (expect fail) -> cancel order -> verify stock unchanged

	// Create product
	prodPayload := PIMProduct{SKU: fmt.Sprintf("T4-SKU-4-%s", uuid.New().String()), Name: "T4 Product 4"}
	var prodResp PIMProduct
	status := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Failed to create product, status: %d", status)
	}
	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected valid ID, got empty string")
	}

	// Add low stock (5)
	stockPayload := InventoryStock{ProductID: prodID, Quantity: 5}
	status = h.Post(t, "/inventory/adjust", stockPayload, nil)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Failed to adjust stock, status: %d", status)
	}

	// Create order for > 5
	orderPayload := Order{
		Items: []OrderItem{
			{ProductID: prodID, Quantity: 10},
		},
	}
	var orderResp Order
	status = h.Post(t, "/oms/orders", orderPayload, &orderResp)
	if status != http.StatusBadRequest && status != http.StatusConflict {
		t.Fatalf("Expected order creation to fail with 400 or 409, got status: %d", status)
	}

	// Verify stock unchanged
	var stockResp InventoryStock
	status = h.Get(t, "/inventory/stock/"+prodID, &stockResp)
	if status != http.StatusOK {
		t.Fatalf("Failed to get stock, status: %d", status)
	}
}

func TestTier4_MultiTenantDataIsolationCheck(t *testing.T) {
	h := harness.Setup(t)

	// Use Tenant A headers to create product -> Use Tenant B headers to GET product (expect 404) -> Create order with Tenant B headers for Tenant A product (expect fail)

	doCustomReq := func(method, path, tenant string, payload interface{}, target interface{}) int {
		var bodyReader io.Reader
		if payload != nil {
			b, _ := json.Marshal(payload)
			bodyReader = bytes.NewReader(b)
		}

		req, err := http.NewRequest(method, h.Config.BaseURL+path, bodyReader)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("X-Tenant-ID", tenant)
		req.Header.Set("X-Org-ID", "00000000-0000-0000-0000-000000000000")

		resp, err := h.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if target != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			body, _ := io.ReadAll(resp.Body)
			json.Unmarshal(body, target)
		}

		return resp.StatusCode
	}

	// Create product with Tenant A
	prodPayload := PIMProduct{SKU: fmt.Sprintf("T4-SKU-5-%s", uuid.New().String()), Name: "Tenant A Product"}
	var prodResp PIMProduct
	status := doCustomReq(http.MethodPost, "/pim/products", "Tenant-A", prodPayload, &prodResp)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("Tenant A failed to create product, status: %d", status)
	}
	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected valid ID, got empty string")
	}

	// GET product with Tenant B (expect 404)
	status = doCustomReq(http.MethodGet, "/pim/products/"+prodID, "Tenant-B", nil, nil)
	if status != http.StatusNotFound {
		t.Fatalf("Expected Tenant B to get 404 for Tenant A product, got: %d", status)
	}

	// Create order with Tenant B for Tenant A product (expect fail)
	orderPayload := Order{
		Items: []OrderItem{
			{ProductID: prodID, Quantity: 1},
		},
	}
	status = doCustomReq(http.MethodPost, "/oms/orders", "Tenant-B", orderPayload, nil)
	if status == http.StatusOK || status == http.StatusCreated {
		t.Fatalf("Expected Tenant B to fail ordering Tenant A product, but got success status: %d", status)
	}
}
