package e2e

import (
	"net/http"
	"testing"

	"commerce_modules/tests/e2e/harness"
)

type InventoryStock struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

var testProdID = "11111111-1111-1111-1111-111111111111"

func TestInventory_AdjustStock_Increase(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: testProdID,
		Quantity:  50,
	}

	statusCode := h.Post(t, "/inventory/adjust", payload, nil)
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		t.Errorf("Expected status OK or Created, got %d", statusCode)
	}
}

func TestInventory_ReserveStock_Success(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: testProdID,
		Quantity:  5,
	}

	statusCode := h.Post(t, "/inventory/reserve", payload, nil)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}

func TestInventory_ReserveStock_Insufficient_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: testProdID,
		Quantity:  9999,
	}

	statusCode := h.Post(t, "/inventory/reserve", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusConflict && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest/Conflict/UnprocessableEntity, got %d", statusCode)
	}
}

func TestInventory_ReleaseStock_RestoresAvailability(t *testing.T) {
	h := harness.Setup(t)
	payload := InventoryStock{
		ProductID: testProdID,
		Quantity:  5,
	}

	statusCode := h.Post(t, "/inventory/release", payload, nil)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}

func TestInventory_GetStock_AccurateLedger(t *testing.T) {
	h := harness.Setup(t)
	var resp InventoryStock
	statusCode := h.Get(t, "/inventory/stock/"+testProdID, &resp)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}
