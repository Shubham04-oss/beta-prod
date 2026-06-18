package unified

import (
	"testing"

	"github.com/google/uuid"
)

func TestMapUnifiedToOrder_VariantIdEmptyFallback(t *testing.T) {
	orgID := uuid.New()
	tenantID := uuid.New()
	productUUID := uuid.New()

	payload := map[string]interface{}{
		"id": "ext-123",
		"line_items": []interface{}{
			map[string]interface{}{
				"variant_id": "",
				"product_id": productUUID.String(),
				"quantity":   float64(1),
				"price":      float64(10.0),
			},
		},
	}

	_, items, err := MapUnifiedToOrder(payload, orgID, tenantID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item")
	}

	if items[0].VariantID == nil {
		t.Fatalf("Bug: VariantID is nil! It failed to fall back to product_id because variant_id was an empty string.")
	}

	if items[0].VariantID.String() != productUUID.String() {
		t.Fatalf("Expected VariantID to be product_id %s, got %s", productUUID.String(), items[0].VariantID.String())
	}
}
