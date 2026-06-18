package oms

import (
	"context"

	"commerce_modules/internal/models"

	"github.com/google/uuid"
)

type Repository interface {
	CreateOrderWithLineItems(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error
	UpdateOrderStatus(ctx context.Context, tenantID, orgID, orderID uuid.UUID, currentStatus, newStatus models.OrderStatus) error
	GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error)
	GetOrderLineItems(ctx context.Context, tenantID, orgID, orderID uuid.UUID) ([]models.OrderLineItem, error)
	CreateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error
	GetCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error
	DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error
}
