package e2e

import (
	"net/http"
	"strings"
	"testing"

	"commerce_modules/tests/e2e/harness"
)

func TestUnifiedSync_PushProduct_EmptyProductID_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := SyncPayload{
		ProductID: "",
	}

	statusCode := h.Post(t, "/unified/sync/push/product", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestUnifiedSync_Webhook_UnknownEvent_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := map[string]interface{}{
		"event": "unknown.event.type",
		"data": map[string]interface{}{
			"product_id": "prod-1",
			"quantity":   100,
		},
	}

	statusCode := h.Post(t, "/unified/webhook", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestUnifiedSync_Webhook_MissingData_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := map[string]interface{}{
		"event": "inventory.updated",
		// data missing
	}

	statusCode := h.Post(t, "/unified/webhook", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestUnifiedSync_PullOrder_MaxOrderIDLength_Succeeds(t *testing.T) {
	h := harness.Setup(t)
	payload := SyncPayload{
		OrderID: strings.Repeat("O", 255), // Max-size input
	}

	statusCode := h.Post(t, "/unified/sync/pull/order", payload, nil)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Errorf("Expected status OK or Created, got %d", statusCode)
	}
}

func TestUnifiedSync_Webhook_MalformedJSON_Fails(t *testing.T) {
	h := harness.Setup(t)

	req, err := http.NewRequest(http.MethodPost, h.Config.BaseURL+"/unified/webhook", strings.NewReader("{malformed: true, [}"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", resp.StatusCode)
	}
}
