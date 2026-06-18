package oms

import (
	"context"
	"fmt"

	"commerce_modules/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OMSService struct {
	repo            Repository
	inventoryClient InventoryClient
	catalogClient   CatalogClient
}

func NewOMSService(repo Repository, inventoryClient InventoryClient, catalogClient CatalogClient) *OMSService {
	return &OMSService{
		repo:            repo,
		inventoryClient: inventoryClient,
		catalogClient:   catalogClient,
	}
}

func (s *OMSService) CreateOrder(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	order.Status = models.OrderStatusPending
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	var totalPrice decimal.Decimal

	for i := range items {
		if items[i].ID == uuid.Nil {
			items[i].ID = uuid.New()
		}

		variant, err := s.catalogClient.GetVariant(ctx, tenantID, orgID, *items[i].VariantID)
		if err != nil {
			return err
		}

		items[i].PriceAtPurchase = variant.Price
		items[i].OptionValuesAtPurchase = variant.OptionValues

		subtotal := variant.Price.Mul(decimal.NewFromInt(int64(items[i].Quantity)))
		totalPrice = totalPrice.Add(subtotal)
	}

	order.TotalPrice = totalPrice

	err := s.repo.CreateOrderWithLineItems(ctx, tenantID, orgID, order, items)
	if err != nil {
		return err
	}

	err = s.inventoryClient.ReserveInventory(ctx, tenantID, orgID, order.ID, items)
	if err != nil {
		rollbackErr := s.repo.UpdateOrderStatus(ctx, tenantID, orgID, order.ID, models.OrderStatusPending, models.OrderStatusFailed)
		if rollbackErr != nil {
			return fmt.Errorf("inventory error: %w, rollback error: %w", err, rollbackErr)
		}
		return fmt.Errorf("inventory error: %w", err)
	}

	return nil
}

func (s *OMSService) FulfillOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) error {
	items, err := s.repo.GetOrderLineItems(ctx, tenantID, orgID, orderID)
	if err != nil {
		return err
	}

	err = s.repo.UpdateOrderStatus(ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusFulfilled)
	if err != nil {
		return err
	}

	err = s.inventoryClient.DeductInventory(ctx, tenantID, orgID, orderID, items)
	if err != nil {
		rollbackErr := s.repo.UpdateOrderStatus(ctx, tenantID, orgID, orderID, models.OrderStatusFulfilled, models.OrderStatusPending)
		if rollbackErr != nil {
			return fmt.Errorf("inventory error: %w, rollback error: %w", err, rollbackErr)
		}
		return fmt.Errorf("inventory error: %w", err)
	}

	return nil
}

func (s *OMSService) CancelOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) error {
	items, err := s.repo.GetOrderLineItems(ctx, tenantID, orgID, orderID)
	if err != nil {
		return err
	}

	err = s.repo.UpdateOrderStatus(ctx, tenantID, orgID, orderID, models.OrderStatusPending, models.OrderStatusCancelled)
	if err != nil {
		return err
	}

	err = s.inventoryClient.ReleaseInventory(ctx, tenantID, orgID, orderID, items)
	if err != nil {
		rollbackErr := s.repo.UpdateOrderStatus(ctx, tenantID, orgID, orderID, models.OrderStatusCancelled, models.OrderStatusPending)
		if rollbackErr != nil {
			return fmt.Errorf("inventory error: %w, rollback error: %w", err, rollbackErr)
		}
		return fmt.Errorf("inventory error: %w", err)
	}

	return nil
}

func (s *OMSService) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	return s.repo.GetOrder(ctx, tenantID, orgID, orderID)
}

func (s *OMSService) CreateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}
	return s.repo.CreateCustomer(ctx, tenantID, orgID, customer)
}

func (s *OMSService) GetCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) (*models.Customer, error) {
	return s.repo.GetCustomer(ctx, tenantID, orgID, customerID)
}

func (s *OMSService) UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	return s.repo.UpdateCustomer(ctx, tenantID, orgID, customer)
}

func (s *OMSService) DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error {
	return s.repo.DeleteCustomer(ctx, tenantID, orgID, customerID)
}
