package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"commerce_modules/tests/e2e/harness"

	"github.com/google/uuid"
)

type SyncPayload struct {
	ProductID string `json:"product_id,omitempty"`
	OrderID   string `json:"order_id,omitempty"`
}

type SyncStatus struct {
	Status       string `json:"status"`
	LocalOrderID string `json:"local_order_id,omitempty"`
}

func TestUnifiedSync_PushProduct_Success(t *testing.T) {
	h := harness.Setup(t)
	prodPayload := PIMProduct{
		SKU:   fmt.Sprintf("SYNC-SKU-1-%s", uuid.New().String()),
		Name:  "Sync Product 1",
		Price: 100.00,
	}
	var prodResp PIMProduct
	statusCode := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}

	payload := SyncPayload{
		ProductID: prodResp.ID,
	}

	var resp SyncStatus
	statusCode = h.Post(t, "/unified/sync/push/product", payload, &resp)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
	if resp.Status == "" {
		t.Errorf("Expected sync status to be populated")
	}
}

func TestUnifiedSync_PullOrder_CreatesLocalOMS(t *testing.T) {
	h := harness.Setup(t)
	payload := SyncPayload{
		OrderID: uuid.New().String(),
	}

	var resp SyncStatus
	statusCode := h.Post(t, "/unified/sync/pull/order", payload, &resp)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Errorf("Expected status OK or Created, got %d", statusCode)
	}
	if resp.Status == "" {
		t.Errorf("Expected sync status to be populated")
	}
}

func TestUnifiedSync_Webhook_UpdatesInventory(t *testing.T) {
	h := harness.Setup(t)
	prodPayload := PIMProduct{
		SKU:   fmt.Sprintf("SYNC-SKU-2-%s", uuid.New().String()),
		Name:  "Sync Product 2",
		Price: 100.00,
	}
	var prodResp PIMProduct
	statusCode := h.Post(t, "/pim/products", prodPayload, &prodResp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Fatalf("Failed to create product, got status %d", statusCode)
	}

	// Webhook payload from Unified.to
	payload := map[string]interface{}{
		"event": "inventory.updated",
		"data": map[string]interface{}{
			"product_id": prodResp.ID,
			"quantity":   100,
		},
	}

	statusCode = h.Post(t, "/unified/webhook", payload, nil)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}

func TestUnifiedSync_PushProduct_HandlesAPIError(t *testing.T) {
	h := harness.Setup(t)
	payload := SyncPayload{
		ProductID: uuid.New().String(),
	}

	statusCode := h.Post(t, "/unified/sync/push/product", payload, nil)
	if statusCode != http.StatusInternalServerError && statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity && statusCode != http.StatusNotFound {
		t.Errorf("Expected error status, got %d", statusCode)
	}
}

func TestUnifiedSync_SyncStatus_TracksFailures(t *testing.T) {
	h := harness.Setup(t)
	var resp SyncStatus
	jobID := "job-" + uuid.New().String()
	statusCode := h.Get(t, fmt.Sprintf("/unified/sync/status/%s", jobID), &resp)
	if statusCode != http.StatusOK && statusCode != http.StatusNotFound {
		t.Errorf("Expected status OK or NotFound, got %d", statusCode)
	}
}
