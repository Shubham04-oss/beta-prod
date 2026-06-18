package unified

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/synq/pkg/db"
	"github.com/unified-to/unified-go-sdk/pkg/models/shared"
)

// MapProductToUnified converts a Synq Product and its Variants into a Unified.to CommerceItem.
func MapProductToUnified(product db.Product, variants []db.ProductVariant) *shared.CommerceItem {
	item := &shared.CommerceItem{
		Name: func(s string) *string { return &s }(product.Title),
	}

	if product.Description.Valid {
		desc := product.Description.String
		item.Description = &desc
	}

	var unifiedVariants []shared.CommerceItemvariant
	for _, v := range variants {
		varID := uuid.UUID(v.ID.Bytes).String()
		variant := shared.CommerceItemvariant{
			ID:   &varID,
			Name: func(s string) *string { return &s }(product.Title), // Fallback
		}
		if v.Sku.Valid {
			variant.Sku = func(s string) *string { return &s }(v.Sku.String)
		}
		if v.Price.Valid {
			// Convert pgtype.Numeric to float64
			f, _ := v.Price.Float64Value()
			if f.Valid {
				variant.Prices = []shared.CommerceItemPrice{
					{
						Price: f.Float64,
					},
				}
			}
		}

		// Map options (e.g., Color, Size) if they exist
		if len(v.OptionValues) > 0 {
			var opts map[string]interface{}
			if err := json.Unmarshal(v.OptionValues, &opts); err == nil {
				var uOpts []shared.CommerceItemOption
				for k, val := range opts {
					uOpts = append(uOpts, shared.CommerceItemOption{
						Name:   k,
						Values: []string{fmt.Sprint(val)},
					})
				}
				variant.Options = uOpts
			}
		}

		unifiedVariants = append(unifiedVariants, variant)
	}

	item.Variants = unifiedVariants

	return item
}

// MapInventoryToUnified converts Synq inventory logic into a Unified.to CommerceInventory.
func MapInventoryToUnified(variantID uuid.UUID, locationID string, availableQuantity float64) *shared.CommerceInventory {
	varIDStr := variantID.String()
	return &shared.CommerceInventory{
		ItemVariantID: &varIDStr,
		LocationID:    func(s string) *string { return &s }(locationID),
		Available:     func(f float64) *float64 { return &f }(availableQuantity),
	}
}
