package oms_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestOMSService_Adversarial_RollbackFailure(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	order := &models.Order{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		OrgID:    uuid.New(),
	}
	items := []models.OrderLineItem{}

	repo.On("CreateOrderWithLineItems", ctx, order.TenantID, order.OrgID, order, items).Return(nil)
	invClient.On("ReserveInventory", ctx, order.TenantID, order.OrgID, mock.AnythingOfType("uuid.UUID"), items).Return(errors.New("inventory reservation failed"))
	repo.On("UpdateOrderStatus", ctx, order.TenantID, order.OrgID, mock.AnythingOfType("uuid.UUID"), models.OrderStatusPending, models.OrderStatusFailed).Return(errors.New("database is down, cannot update status"))

	err := svc.CreateOrder(ctx, order.TenantID, order.OrgID, order, items)
	if err == nil {
		t.Fatalf("Expected error due to inventory failure")
	}

	if !strings.Contains(err.Error(), "inventory reservation failed") || !strings.Contains(err.Error(), "database is down") {
		t.Errorf("Expected combined errors, got %v", err)
	}

	repo.AssertExpectations(t)
	invClient.AssertExpectations(t)
	catalogClient.AssertExpectations(t)
}
