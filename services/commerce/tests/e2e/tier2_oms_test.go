package e2e

import (
	"net/http"
	"testing"

	"commerce_modules/tests/e2e/harness"
)

func TestOMS_CreateOrder_EmptyItems_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := Order{
		Items: []OrderItem{},
	}

	statusCode := h.Post(t, "/oms/orders", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestOMS_CreateOrder_ZeroQuantity_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := Order{
		Items: []OrderItem{
			{ProductID: "prod-1", Quantity: 0},
		},
	}

	statusCode := h.Post(t, "/oms/orders", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestOMS_CreateOrder_NegativeQuantity_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := Order{
		Items: []OrderItem{
			{ProductID: "prod-1", Quantity: -5},
		},
	}

	statusCode := h.Post(t, "/oms/orders", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestOMS_UpdateOrderStatus_InvalidStateTransition_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := map[string]string{
		"status": "shipped",
	}

	// Assuming ord-cancelled is already cancelled, or we try invalid transition.
	// We'll just target an order and assume transition from cancelled to shipped is invalid.
	statusCode := h.Put(t, "/oms/orders/ord-cancelled/status", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusConflict {
		t.Errorf("Expected status BadRequest or Conflict, got %d", statusCode)
	}
}

func TestOMS_CreateOrder_MaxItemsLimit_FailsOrSucceeds(t *testing.T) {
	h := harness.Setup(t)

	// Create massive items slice
	items := make([]OrderItem, 5000)
	for i := 0; i < 5000; i++ {
		items[i] = OrderItem{ProductID: "prod-1", Quantity: 1}
	}

	payload := Order{
		Items: items,
	}

	statusCode := h.Post(t, "/oms/orders", payload, nil)
	if statusCode != http.StatusRequestEntityTooLarge && statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Errorf("Expected status RequestEntityTooLarge, Created, or OK, got %d", statusCode)
	}
}
