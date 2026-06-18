package e2e

import (
	"commerce_modules/tests/e2e/harness"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

type PIMProduct struct {
	ID          string  `json:"id,omitempty"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func TestPIM_CreateProduct_Success(t *testing.T) {
	h := harness.Setup(t)
	payload := PIMProduct{
		SKU:         fmt.Sprintf("SKU-12345-%s", uuid.New().String()),
		Name:        "Test Product",
		Description: "A product for E2E testing",
		Price:       99.99,
	}

	var resp PIMProduct
	statusCode := h.Post(t, "/pim/products", payload, &resp)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Errorf("Expected status Created or OK, got %d", statusCode)
	}
}

func TestPIM_UpdateProduct_ModifyAttributes(t *testing.T) {
	h := harness.Setup(t)
	// Create first
	createPayload := PIMProduct{SKU: fmt.Sprintf("SKU-UPDATE-1-%s", uuid.New().String()), Name: "Original", Price: 10.0}
	var created PIMProduct
	h.Post(t, "/pim/products", createPayload, &created)

	payload := PIMProduct{
		Name:  "Updated Product Name",
		Price: 89.99,
	}

	var resp PIMProduct
	statusCode := h.Put(t, "/pim/products/"+created.ID, payload, &resp)
	if statusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", statusCode)
	}
}

func TestPIM_GetProduct_NotFound(t *testing.T) {
	h := harness.Setup(t)
	var resp PIMProduct
	statusCode := h.Get(t, "/pim/products/non-existent-id", &resp)
	if statusCode != http.StatusNotFound {
		t.Errorf("Expected status NotFound, got %d", statusCode)
	}
}

func TestPIM_DeleteProduct_Success(t *testing.T) {
	h := harness.Setup(t)
	// Create first
	createPayload := PIMProduct{SKU: fmt.Sprintf("SKU-DELETE-1-%s", uuid.New().String()), Name: "ToDelete", Price: 10.0}
	var created PIMProduct
	h.Post(t, "/pim/products", createPayload, &created)

	statusCode := h.Delete(t, "/pim/products/"+created.ID)
	if statusCode != http.StatusNoContent && statusCode != http.StatusOK {
		t.Errorf("Expected status NoContent or OK, got %d", statusCode)
	}
}

func TestPIM_CreateProduct_DuplicateSKU_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := PIMProduct{
		SKU:         fmt.Sprintf("SKU-DUPLICATE-%s", uuid.New().String()),
		Name:        "Duplicate SKU Product",
		Description: "Should fail on second create",
		Price:       10.00,
	}

	// First attempt
	h.Post(t, "/pim/products", payload, nil)

	// Second attempt should fail
	statusCode := h.Post(t, "/pim/products", payload, nil)
	if statusCode != http.StatusConflict && statusCode != http.StatusBadRequest {
		t.Errorf("Expected status Conflict or BadRequest, got %d", statusCode)
	}
}
