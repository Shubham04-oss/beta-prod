package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"commerce_modules/tests/e2e/harness"

	"github.com/google/uuid"
)

func TestTier3_PIM_OMS_ProductOrderLifecycle(t *testing.T) {
	h := harness.Setup(t)

	// 1. PIM: POST /pim/products
	prodPayload := PIMProduct{
		SKU:         fmt.Sprintf("TIER3-SKU-1-%s", uuid.New().String()),
		Name:        "Tier 3 Product",
		Description: "Product for PIM + OMS test",
		Price:       150.00,
	}
	var prodResp PIMProduct
	statusCode := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}

	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected non-empty product ID")
	}

	// 1.5. Inventory: POST /inventory/adjust
	stockPayload := InventoryStock{
		ProductID: prodID,
		Quantity:  10,
	}
	statusCode = h.Post(t, "/inventory/adjust", stockPayload, nil)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Fatalf("Failed to add stock, got status %d", statusCode)
	}

	// 2. OMS: POST /oms/orders
	orderPayload := Order{
		Items: []OrderItem{
			{ProductID: prodID, Quantity: 1},
		},
	}
	var orderResp Order
	statusCode = h.Post(t, "/oms/orders", orderPayload, &orderResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create order, got status %d", statusCode)
	}

	orderID := orderResp.ID
	if orderID == "" {
		t.Fatalf("Expected non-empty order ID")
	}

	// 3. OMS: GET /oms/orders/{id}
	var getOrderResp Order
	statusCode = h.Get(t, fmt.Sprintf("/oms/orders/%s", orderID), &getOrderResp)
	if statusCode != http.StatusOK {
		t.Fatalf("Expected status OK for GET order, got %d", statusCode)
	}
	if getOrderResp.ID != orderID {
		t.Fatalf("Expected order ID %s, got %s", orderID, getOrderResp.ID)
	}
}

func TestTier3_PIM_Inventory_ProductStockSetup(t *testing.T) {
	h := harness.Setup(t)

	// 1. PIM: POST /pim/products
	prodPayload := PIMProduct{
		SKU:   fmt.Sprintf("TIER3-SKU-2-%s", uuid.New().String()),
		Name:  "Tier 3 Product 2",
		Price: 200.00,
	}
	var prodResp PIMProduct
	statusCode := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}

	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected non-empty product ID")
	}

	// 2. Inventory: POST /inventory/adjust
	stockPayload := InventoryStock{
		ProductID: prodID,
		Quantity:  100,
	}
	statusCode = h.Post(t, "/inventory/adjust", stockPayload, nil)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Fatalf("Failed to adjust stock, got status %d", statusCode)
	}

	// 3. Inventory: GET /inventory/stock/{id}
	var stockResp InventoryStock
	statusCode = h.Get(t, fmt.Sprintf("/inventory/stock/%s", prodID), &stockResp)
	if statusCode != http.StatusOK {
		t.Fatalf("Expected status OK for GET stock, got %d", statusCode)
	}
	if stockResp.Quantity != 100 {
		t.Fatalf("Expected stock quantity 100, got %d", stockResp.Quantity)
	}
}

func TestTier3_PIM_UnifiedSync_ProductExport(t *testing.T) {
	h := harness.Setup(t)

	// 1. PIM: POST /pim/products
	prodPayload := PIMProduct{
		SKU:   fmt.Sprintf("TIER3-SKU-3-%s", uuid.New().String()),
		Name:  "Tier 3 Product 3",
		Price: 300.00,
	}
	var prodResp PIMProduct
	statusCode := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}

	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected non-empty product ID")
	}

	// 2. Unified Sync: POST /unified/sync/push/product
	syncPayload := SyncPayload{
		ProductID: prodID,
	}
	var syncResp SyncStatus
	statusCode = h.Post(t, "/unified/sync/push/product", syncPayload, &syncResp)
	if statusCode != http.StatusOK {
		t.Fatalf("Failed to push product, got status %d", statusCode)
	}
	if syncResp.Status == "" {
		t.Fatalf("Expected sync status to be populated")
	}

	// No job ID returned by API, cannot verify sync status directly here without faking.
}

func TestTier3_OMS_Inventory_OrderFulfillmentFlow(t *testing.T) {
	h := harness.Setup(t)

	prodPayload := PIMProduct{
		SKU:   fmt.Sprintf("TIER3-SKU-4-%s", uuid.New().String()),
		Name:  "Tier 3 Product 4",
		Price: 400.00,
	}
	var prodResp PIMProduct
	statusCode := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}
	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected non-empty product ID")
	}

	// 1. Inventory: POST /inventory/adjust
	stockPayload := InventoryStock{
		ProductID: prodID,
		Quantity:  50,
	}
	statusCode = h.Post(t, "/inventory/adjust", stockPayload, nil)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Fatalf("Failed to adjust stock, got status %d", statusCode)
	}

	// 2. OMS: POST /oms/orders
	orderPayload := Order{
		Items: []OrderItem{
			{ProductID: prodID, Quantity: 5},
		},
	}
	var orderResp Order
	statusCode = h.Post(t, "/oms/orders", orderPayload, &orderResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create order, got status %d", statusCode)
	}

	orderID := orderResp.ID
	if orderID == "" {
		t.Fatalf("Expected non-empty order ID")
	}

	// 3. Inventory: GET /inventory/stock/{id}
	var postOrderStock InventoryStock
	statusCode = h.Get(t, fmt.Sprintf("/inventory/stock/%s", prodID), &postOrderStock)
	if statusCode != http.StatusOK {
		t.Fatalf("Expected status OK for GET stock, got %d", statusCode)
	}
	if postOrderStock.Quantity != 45 {
		t.Fatalf("Expected stock quantity 45 after order creation, got %d", postOrderStock.Quantity)
	}

	// 4. OMS: PUT /oms/orders/{id}/status
	statusPayload := map[string]string{
		"status": "cancelled",
	}
	statusCode = h.Put(t, fmt.Sprintf("/oms/orders/%s/status", orderID), statusPayload, nil)
	if statusCode != http.StatusOK {
		t.Fatalf("Failed to update order status, got status %d", statusCode)
	}

	// 5. Inventory: GET /inventory/stock/{id}
	var stockResp InventoryStock
	statusCode = h.Get(t, fmt.Sprintf("/inventory/stock/%s", prodID), &stockResp)
	if statusCode != http.StatusOK {
		t.Fatalf("Expected status OK for GET stock, got %d", statusCode)
	}
	if stockResp.Quantity != 50 {
		t.Fatalf("Expected stock quantity 50 after cancellation, got %d", stockResp.Quantity)
	}
}

func TestTier3_OMS_UnifiedSync_OrderImport(t *testing.T) {
	h := harness.Setup(t)

	remoteOrderID := uuid.New().String()

	// 1. Unified Sync: POST /unified/sync/pull/order
	syncPayload := SyncPayload{
		OrderID: remoteOrderID,
	}
	var syncResp SyncStatus
	statusCode := h.Post(t, "/unified/sync/pull/order", syncPayload, &syncResp)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Fatalf("Failed to pull order, got status %d", statusCode)
	}
	if syncResp.Status == "" {
		t.Fatalf("Expected sync status to be populated")
	}

	localOrderID := syncResp.LocalOrderID
	if localOrderID == "" {
		t.Fatalf("Expected non-empty local order ID from pull response")
	}

	// 2. OMS: GET /oms/orders/{id}
	var orderResp Order
	statusCode = h.Get(t, fmt.Sprintf("/oms/orders/%s", localOrderID), &orderResp)
	if statusCode != http.StatusOK {
		t.Fatalf("GET order returned status %d", statusCode)
	}
	if orderResp.ID != localOrderID {
		t.Fatalf("Expected order ID %s, got %s", localOrderID, orderResp.ID)
	}

	// 3. OMS: PUT /oms/orders/{id}/status
	statusPayload := map[string]string{
		"status": "shipped",
	}
	statusCode = h.Put(t, fmt.Sprintf("/oms/orders/%s/status", localOrderID), statusPayload, nil)
	if statusCode != http.StatusOK {
		t.Fatalf("Failed to update order status, got status %d", statusCode)
	}
}

func TestTier3_Inventory_UnifiedSync_ExternalStockUpdate(t *testing.T) {
	h := harness.Setup(t)

	prodPayload := PIMProduct{
		SKU:   fmt.Sprintf("TIER3-SKU-6-%s", uuid.New().String()),
		Name:  "Tier 3 Product 6",
		Price: 600.00,
	}
	var prodResp PIMProduct
	statusCode := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}
	prodID := prodResp.ID
	if prodID == "" {
		t.Fatalf("Expected non-empty product ID")
	}

	// 1. Unified Sync: POST /unified/webhook
	webhookPayload := map[string]interface{}{
		"event": "inventory.updated",
		"data": map[string]interface{}{
			"product_id": prodID,
			"quantity":   150,
		},
	}
	statusCode = h.Post(t, "/unified/webhook", webhookPayload, nil)
	if statusCode != http.StatusOK {
		t.Fatalf("Failed to process webhook, got status %d", statusCode)
	}

	// 2. Inventory: GET /inventory/stock/{id}
	var stockResp InventoryStock
	statusCode = h.Get(t, fmt.Sprintf("/inventory/stock/%s", prodID), &stockResp)
	if statusCode != http.StatusOK {
		t.Fatalf("Expected status OK for GET stock, got %d", statusCode)
	}
	if stockResp.Quantity != 150 {
		t.Fatalf("Expected stock quantity 150, got %d", stockResp.Quantity)
	}
}
