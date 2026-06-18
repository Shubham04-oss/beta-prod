package oms_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce_modules/internal/models"
	"commerce_modules/internal/oms"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestOMSService_FulfillOrder_RaceCondition(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()
	orderID := uuid.New()

	var deductions int32

	var currentStatus models.OrderStatus = models.OrderStatusPending
	var mu sync.Mutex

	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusFulfilled).Return(nil).Run(func(args mock.Arguments) {
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		if currentStatus != models.OrderStatusPending {
			panic("order not found or in invalid state") // simulate error
		}
		currentStatus = models.OrderStatusFulfilled
	})

	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return([]models.OrderLineItem{}, nil)

	invClient.On("DeductInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(nil).Run(func(args mock.Arguments) {
		mu.Lock()
		deductions++
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { recover() }() // catch the panic from race
			_ = svc.FulfillOrder(ctx, tenantID, orgID, orderID)
		}()
	}

	wg.Wait()

	if deductions > 1 {
		t.Fatalf("Race condition detected: DeductInventory called %d times for a single order!", deductions)
	}
	catalogClient.AssertExpectations(t)
}

func TestOMSService_CreateOrder_IncompleteRollback(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	order := &models.Order{
		TenantID: uuid.New(),
		OrgID:    uuid.New(),
	}
	items := []models.OrderLineItem{}

	repo.On("CreateOrderWithLineItems", ctx, order.TenantID, order.OrgID, order, items).Return(nil)
	invClient.On("ReserveInventory", ctx, order.TenantID, order.OrgID, mock.AnythingOfType("uuid.UUID"), items).Return(errors.New("inventory reservation failed"))

	updateCalled := false
	repo.On("UpdateOrderStatus", ctx, order.TenantID, order.OrgID, mock.AnythingOfType("uuid.UUID"), models.OrderStatusPending, models.OrderStatusFailed).Return(nil).Run(func(args mock.Arguments) {
		updateCalled = true
	})

	err := svc.CreateOrder(ctx, order.TenantID, order.OrgID, order, items)
	if err == nil {
		t.Fatalf("Expected error due to inventory failure")
	}

	if !updateCalled {
		t.Fatalf("Rollback to FAILED was not performed!")
	}
	catalogClient.AssertExpectations(t)
}

func TestOMSService_Adversarial_PartialFailures(t *testing.T) {
	repo := &mockRepository{}
	invClient := &mockInventoryClient{}
	catalogClient := &mockCatalogClient{}
	svc := oms.NewOMSService(repo, invClient, catalogClient)

	ctx := context.Background()
	tenantID := uuid.New()
	orgID := uuid.New()
	orderID := uuid.New()

	var currentStatus models.OrderStatus = models.OrderStatusPending
	var mu sync.Mutex

	repo.On("UpdateOrderStatus", ctx, tenantID, orgID, orderID, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		curStatus := args.Get(4).(models.OrderStatus)
		newStatus := args.Get(5).(models.OrderStatus)
		mu.Lock()
		defer mu.Unlock()
		if currentStatus != curStatus {
			panic("invalid state transition")
		}
		currentStatus = newStatus
	})

	repo.On("GetOrderLineItems", ctx, tenantID, orgID, orderID).Return([]models.OrderLineItem{}, nil)

	invClient.On("DeductInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(errors.New("timeout connecting to inventory")).Once()
	invClient.On("DeductInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(nil)

	invClient.On("ReleaseInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(errors.New("inventory lock acquisition failed")).Once()
	invClient.On("ReleaseInventory", ctx, tenantID, orgID, orderID, []models.OrderLineItem{}).Return(nil)

	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		return svc.FulfillOrder(ctx, tenantID, orgID, orderID)
	}()
	if err == nil {
		t.Fatalf("Expected error from FulfillOrder")
	}

	mu.Lock()
	if currentStatus != models.OrderStatusPending {
		t.Fatalf("Expected status to be rolled back to PENDING, got %s", currentStatus)
	}
	mu.Unlock()

	err = func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		return svc.CancelOrder(ctx, tenantID, orgID, orderID)
	}()
	if err == nil {
		t.Fatalf("Expected error from CancelOrder")
	}

	mu.Lock()
	if currentStatus != models.OrderStatusPending {
		t.Fatalf("Expected status to be rolled back to PENDING, got %s", currentStatus)
	}
	mu.Unlock()
	catalogClient.AssertExpectations(t)
}
