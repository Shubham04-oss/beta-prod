package e2e

import (
	"math"
	"net/http"
	"testing"

	"commerce_modules/tests/e2e/harness"
)

func TestInventory_AdjustStock_ZeroQuantity_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: "prod-1",
		Quantity:  0,
	}

	statusCode := h.Post(t, "/inventory/adjust", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestInventory_AdjustStock_EmptyProductID_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: "",
		Quantity:  10,
	}

	statusCode := h.Post(t, "/inventory/adjust", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestInventory_ReserveStock_NegativeQuantity_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: "prod-1",
		Quantity:  -5,
	}

	statusCode := h.Post(t, "/inventory/reserve", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestInventory_ReleaseStock_MoreThanReserved_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: "prod-1",
		Quantity:  1000000, // Assuming this is more than reserved
	}

	statusCode := h.Post(t, "/inventory/release", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusConflict && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest, Conflict, or UnprocessableEntity, got %d", statusCode)
	}
}

func TestInventory_AdjustStock_MaxIntQuantity_OverflowCheck(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: "prod-1",
		Quantity:  math.MaxInt, // Massive quantity to check for overflow handling
	}

	statusCode := h.Post(t, "/inventory/adjust", payload, nil)
	// It might succeed if big ints are supported, or fail if bounded.
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity && statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Errorf("Expected status BadRequest/UnprocessableEntity or OK/Created, got %d", statusCode)
	}
}
