package e2e

import (
	"bytes"
	"commerce_modules/tests/e2e/harness"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPIM_CreateProduct_EmptySKU_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := PIMProduct{
		SKU:         "",
		Name:        "Test Product Empty SKU",
		Description: "A product with empty SKU",
		Price:       10.0,
	}

	statusCode := h.Post(t, "/pim/products", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestPIM_CreateProduct_NegativePrice_Fails(t *testing.T) {
	h := harness.Setup(t)
	payload := PIMProduct{
		SKU:         fmt.Sprintf("SKU-NEG-PRICE-%s", uuid.New().String()),
		Name:        "Test Product Negative Price",
		Description: "A product with negative price",
		Price:       -10.5,
	}

	statusCode := h.Post(t, "/pim/products", payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestPIM_CreateProduct_MaxNameLength_Succeeds(t *testing.T) {
	h := harness.Setup(t)
	payload := PIMProduct{
		SKU:         fmt.Sprintf("SKU-MAX-NAME-%s", uuid.New().String()),
		Name:        strings.Repeat("A", 255),
		Description: "A product with exactly 255 chars in name",
		Price:       99.99,
	}

	statusCode := h.Post(t, "/pim/products", payload, nil)
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		t.Errorf("Expected status Created or OK, got %d", statusCode)
	}
}

func TestPIM_UpdateProduct_EmptyPayload_Fails(t *testing.T) {
	h := harness.Setup(t)
	// Create first
	createPayload := PIMProduct{SKU: fmt.Sprintf("SKU-EMPTY-UPDATE-%s", uuid.New().String()), Name: "Original", Price: 10.0}
	var created PIMProduct
	h.Post(t, "/pim/products", createPayload, &created)

	payload := map[string]interface{}{}

	statusCode := h.Put(t, "/pim/products/"+created.ID, payload, nil)
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", statusCode)
	}
}

func TestPIM_CreateProduct_MalformedJSON_Fails(t *testing.T) {
	h := harness.Setup(t)

	req, _ := http.NewRequest("POST", h.Config.BaseURL+"/pim/products", bytes.NewBufferString("{invalid_json: true}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Org-ID", "00000000-0000-0000-0000-000000000000")
	req.Header.Set("X-Tenant-ID", "00000000-0000-0000-0000-000000000000")
	resp, err := h.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status BadRequest or UnprocessableEntity, got %d", resp.StatusCode)
	}
}
