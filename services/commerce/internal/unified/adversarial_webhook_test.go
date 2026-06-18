package unified

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type AdversarialInventoryClient struct {
	MockInventoryClient
	getStockErr error
	stock       map[string]int
}

func (m *AdversarialInventoryClient) GetStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string) (int, error) {
	if m.getStockErr != nil {
		return 0, m.getStockErr
	}
	return m.stock[productID], nil
}

func (m *AdversarialInventoryClient) AdjustStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string, quantity int) error {
	m.stock[productID] += quantity
	return nil
}

func TestHandleWebhook_InitialStock_NotFoundBug(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &MockOMSClient{}

	// Simulate "not found" when GetStock is called for a new product
	inv := &AdversarialInventoryClient{
		getStockErr: errors.New("inventory record not found"),
		stock:       make(map[string]int),
	}

	svc := NewSyncService(uc, pim, oms, inv)

	productID := "new-prod-id"

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"product_id": productID,
			"quantity":   float64(50),
		},
	}

	err := svc.HandleWebhook(context.Background(), "inventory.updated", payload)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Fatalf("Bug: HandleWebhook fails to process inventory update for new products because GetStock returns 'not found'. It should treat 'not found' as 0 stock and adjust it. Error: %v", err)
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	if inv.stock[productID] != 50 {
		t.Fatalf("Expected stock to be 50, got %d", inv.stock[productID])
	}
}
