package tests

import (
	"context"
	"testing"

	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOMS_StateMachine_Race(t *testing.T) {
	// This test just shows that FulfillOrder doesn't check the current status.
	// We'll mock the repo and inventory client to succeed.
	repo := new(MockRepo)
	inv := new(MockInventoryClient)
	cat := new(MockCatalogClient)
	svc := oms.NewOMSService(repo, inv, cat)

	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()
	orderID := uuid.New()

	items := []models.OrderLineItem{}

	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusFulfilled).Return(nil)
	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return(items, nil)
	inv.On("DeductInventory", ctx, tenantID, orgID, orderID, items).Return(nil)

	err := svc.FulfillOrder(ctx, tenantID, orgID, orderID)
	assert.NoError(t, err)

	// In a real implementation, if the order is already CANCELLED,
	// FulfillOrder shouldn't succeed. The repository should use something like:
	// UPDATE orders SET status = $1 WHERE id = $2 AND status = 'PENDING'
	cat.AssertExpectations(t)
}
