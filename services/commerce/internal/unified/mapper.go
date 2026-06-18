package unified

import (
	"commerce_modules/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MapProductToUnified converts a product and its variants to the Unified.to commerce product payload
func MapProductToUnified(product *models.Product, variants []*models.ProductVariant) map[string]interface{} {
	payload := map[string]interface{}{
		"external_id": product.ID.String(),
		"name":        product.Title,
	}

	if product.Description != nil {
		payload["description"] = *product.Description
	}

	unifiedVariants := make([]map[string]interface{}, 0, len(variants))
	for _, v := range variants {
		vMap := map[string]interface{}{
			"external_id": v.ID.String(),
			"price":       v.Price.InexactFloat64(),
			"currency":    v.Currency,
		}
		if v.SKU != nil {
			vMap["sku"] = *v.SKU
		}
		if v.Barcode != nil {
			vMap["barcode"] = *v.Barcode
		}
		unifiedVariants = append(unifiedVariants, vMap)
	}

	payload["variants"] = unifiedVariants
	return payload
}

// MapUnifiedToOrder converts a Unified.to commerce order payload to our Order model
func MapUnifiedToOrder(payload map[string]interface{}, orgID, tenantID uuid.UUID) (*models.Order, []models.OrderLineItem, error) {
	var orderID uuid.UUID
	extID := ""
	if idVal, ok := payload["id"].(string); ok && idVal != "" {
		extID = idVal
	} else if extIdVal, ok := payload["external_id"].(string); ok && extIdVal != "" {
		extID = extIdVal
	}

	if extID != "" {
		orderID = uuid.NewMD5(uuid.NameSpaceOID, []byte(extID))
	} else {
		orderID = uuid.New()
	}

	order := &models.Order{
		ID:        orderID,
		OrgID:     orgID,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if status, ok := payload["status"].(string); ok {
		switch status {
		case "PAID":
			order.Status = models.OrderStatusPaid
		case "PENDING":
			order.Status = models.OrderStatusPending
		case "FULFILLED":
			order.Status = models.OrderStatusFulfilled
		case "CANCELLED":
			order.Status = models.OrderStatusCancelled
		default:
			order.Status = models.OrderStatusPending
		}
	} else {
		order.Status = models.OrderStatusPending
	}

	if currency, ok := payload["currency"].(string); ok {
		order.Currency = currency
	} else {
		order.Currency = "USD"
	}

	if totalPrice, ok := payload["total_price"].(float64); ok {
		order.TotalPrice = decimal.NewFromFloat(totalPrice)
	} else if totalPriceStr, ok := payload["total_price"].(string); ok {
		if d, err := decimal.NewFromString(totalPriceStr); err == nil {
			order.TotalPrice = d
		}
	}

	var lineItems []models.OrderLineItem
	if items, ok := payload["line_items"].([]interface{}); ok {
		for _, itemIf := range items {
			itemMap, ok := itemIf.(map[string]interface{})
			if !ok {
				continue
			}

			li := &models.OrderLineItem{
				ID:        uuid.New(),
				OrgID:     orgID,
				TenantID:  tenantID,
				OrderID:   order.ID,
				CreatedAt: time.Now(),
			}

			if qty, ok := itemMap["quantity"].(float64); ok {
				li.Quantity = int(qty)
			}
			if price, ok := itemMap["price"].(float64); ok {
				li.PriceAtPurchase = decimal.NewFromFloat(price)
			}
			if vidStr, ok := itemMap["variant_id"].(string); ok && vidStr != "" {
				if vid, err := uuid.Parse(vidStr); err == nil {
					li.VariantID = &vid
				}
			} else if pidStr, ok := itemMap["product_id"].(string); ok && pidStr != "" {
				// Fallback to product ID if variant ID is not present
				if pid, err := uuid.Parse(pidStr); err == nil {
					li.VariantID = &pid
				}
			}

			lineItems = append(lineItems, *li)
		}
	} else if items, ok := payload["line_items"].([]map[string]interface{}); ok {
		// Type assertion for specifically typed slices
		for _, itemMap := range items {
			li := &models.OrderLineItem{
				ID:        uuid.New(),
				OrgID:     orgID,
				TenantID:  tenantID,
				OrderID:   order.ID,
				CreatedAt: time.Now(),
			}

			if qty, ok := itemMap["quantity"].(float64); ok {
				li.Quantity = int(qty)
			}
			if price, ok := itemMap["price"].(float64); ok {
				li.PriceAtPurchase = decimal.NewFromFloat(price)
			}
			if vidStr, ok := itemMap["variant_id"].(string); ok && vidStr != "" {
				if vid, err := uuid.Parse(vidStr); err == nil {
					li.VariantID = &vid
				}
			} else if pidStr, ok := itemMap["product_id"].(string); ok && pidStr != "" {
				if pid, err := uuid.Parse(pidStr); err == nil {
					li.VariantID = &pid
				}
			}

			lineItems = append(lineItems, *li)
		}
	}

	return order, lineItems, nil
}

// ExtractInventoryUpdate extracts inventory updates from a webhook payload
func ExtractInventoryUpdate(payload map[string]interface{}) (string, int, error) {
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return "", 0, fmt.Errorf("missing data object")
	}

	productID, ok := data["product_id"].(string)
	if !ok || productID == "" {
		return "", 0, fmt.Errorf("missing product_id")
	}

	quantityFloat, ok := data["quantity"].(float64)
	if !ok {
		return "", 0, fmt.Errorf("missing or invalid quantity")
	}

	return productID, int(quantityFloat), nil
}
