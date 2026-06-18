package unified

import (
	"testing"

	"github.com/google/uuid"
)

func TestMapUnifiedToOrder_Adversarial(t *testing.T) {
	orgID := uuid.New()
	tenantID := uuid.New()

	payload := map[string]interface{}{
		"id":          "", // empty id
		"external_id": "", // empty external_id
		"line_items": []interface{}{
			map[string]interface{}{
				"variant_id": "",             // empty variant_id
				"product_id": "invalid-uuid", // invalid uuid for product_id
			},
			map[string]interface{}{
				"variant_id": 123, // wrong type
				"product_id": uuid.New().String(),
			},
		},
	}

	order, items, err := MapUnifiedToOrder(payload, orgID, tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order == nil {
		t.Fatalf("expected order")
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// First item: invalid product_id UUID -> VariantID should be nil
	if items[0].VariantID != nil {
		t.Errorf("expected nil VariantID for item 0, got %s", items[0].VariantID.String())
	}

	// Second item: wrong type for variant_id, fallback to product_id
	if items[1].VariantID == nil {
		t.Errorf("expected non-nil VariantID for item 1")
	}
}
