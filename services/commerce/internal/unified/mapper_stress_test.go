package unified

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestMapUnifiedToOrder_EdgeCases(t *testing.T) {
	orgID := uuid.New()
	tenantID := uuid.New()

	// 1. Missing or string-based TotalPrice
	payload1 := map[string]interface{}{
		"status":      "PAID",
		"total_price": "100.50", // Sent as string
		"line_items": []interface{}{
			map[string]interface{}{
				"quantity":   float64(2),
				"price":      float64(50.25),
				"variant_id": uuid.New().String(),
			},
		},
	}

	order1, _, err := MapUnifiedToOrder(payload1, orgID, tenantID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if order1.TotalPrice.Cmp(decimal.NewFromFloat(0)) == 0 {
		t.Errorf("TotalPrice is 0, string to decimal conversion failed because it strictly checks for float64.")
	}

	// 2. Duplicate line items if both line_items and items exist
	payload2 := map[string]interface{}{
		"status": "PAID",
		"line_items": []interface{}{
			map[string]interface{}{
				"quantity": float64(1),
				"price":    float64(10),
			},
		},
		"items": []interface{}{
			map[string]interface{}{
				"quantity": float64(1),
			},
		},
	}

	_, items2, err := MapUnifiedToOrder(payload2, orgID, tenantID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items2) == 2 {
		t.Errorf("Duplicate line items created: processed both 'line_items' and 'items'")
	}
}
