package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusFulfilled OrderStatus = "FULFILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type UserRole string

const (
	UserRoleAdmin   UserRole = "ADMIN"
	UserRoleManager UserRole = "MANAGER"
	UserRoleAgent   UserRole = "AGENT"
)

type Organization struct {
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at"`
	Name      string          `json:"name"`
	Metadata  json.RawMessage `json:"metadata"`
	ID        uuid.UUID       `json:"id"`
}

type Tenant struct {
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at"`
	Name      string          `json:"name"`
	Metadata  json.RawMessage `json:"metadata"`
	ID        uuid.UUID       `json:"id"`
	OrgID     uuid.UUID       `json:"org_id"`
}

type User struct {
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	TenantID  *uuid.UUID      `json:"tenant_id"`
	DeletedAt *time.Time      `json:"deleted_at"`
	Email     string          `json:"email"`
	Role      UserRole        `json:"role"`
	Metadata  json.RawMessage `json:"metadata"`
	ID        uuid.UUID       `json:"id"`
	OrgID     uuid.UUID       `json:"org_id"`
}

type Product struct {
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	UpdatedBy   *uuid.UUID       `json:"updated_by"`
	Description *string          `json:"description"`
	DeletedAt   *time.Time       `json:"deleted_at"`
	CreatedBy   *uuid.UUID       `json:"created_by"`
	Embedding   *pgvector.Vector `json:"embedding"`
	Title       string           `json:"title"`
	Status      string           `json:"status"`
	Options     json.RawMessage  `json:"options"`
	Metadata    json.RawMessage  `json:"metadata"`
	ID          uuid.UUID        `json:"id"`
	OrgID       uuid.UUID        `json:"org_id"`
	TenantID    uuid.UUID        `json:"tenant_id"`
}

type ProductVariant struct {
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	UpdatedBy    *uuid.UUID      `json:"updated_by"`
	SKU          *string         `json:"sku"`
	DeletedAt    *time.Time      `json:"deleted_at"`
	CreatedBy    *uuid.UUID      `json:"created_by"`
	Barcode      *string         `json:"barcode"`
	Currency     string          `json:"currency"`
	Price        decimal.Decimal `json:"price"`
	OptionValues json.RawMessage `json:"option_values"`
	Metadata     json.RawMessage `json:"metadata"`
	ID           uuid.UUID       `json:"id"`
	OrgID        uuid.UUID       `json:"org_id"`
	TenantID     uuid.UUID       `json:"tenant_id"`
	ProductID    uuid.UUID       `json:"product_id"`
}

type Location struct {
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Metadata  json.RawMessage `json:"metadata"`
	ID        uuid.UUID       `json:"id"`
	OrgID     uuid.UUID       `json:"org_id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
}

type InventoryLevel struct {
	UpdatedBy         *uuid.UUID `json:"updated_by"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ID                uuid.UUID  `json:"id"`
	OrgID             uuid.UUID  `json:"org_id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	VariantID         uuid.UUID  `json:"variant_id"`
	LocationID        uuid.UUID  `json:"location_id"`
	AvailableQuantity int        `json:"available_quantity"`
	ReservedQuantity  int        `json:"reserved_quantity"`
}

type Customer struct {
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	FirstName *string         `json:"first_name"`
	LastName  *string         `json:"last_name"`
	DeletedAt *time.Time      `json:"deleted_at"`
	Email     string          `json:"email"`
	Metadata  json.RawMessage `json:"metadata"`
	ID        uuid.UUID       `json:"id"`
	OrgID     uuid.UUID       `json:"org_id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
}

type Order struct {
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	CustomerID *uuid.UUID      `json:"customer_id"`
	DeletedAt  *time.Time      `json:"deleted_at"`
	Currency   string          `json:"currency"`
	Status     OrderStatus     `json:"status"`
	TotalPrice decimal.Decimal `json:"total_price"`
	Metadata   json.RawMessage `json:"metadata"`
	ID         uuid.UUID       `json:"id"`
	OrgID      uuid.UUID       `json:"org_id"`
	TenantID   uuid.UUID       `json:"tenant_id"`
}

type OrderLineItem struct {
	CreatedAt              time.Time       `json:"created_at"`
	VariantID              *uuid.UUID      `json:"variant_id"`
	PriceAtPurchase        decimal.Decimal `json:"price_at_purchase"`
	OptionValuesAtPurchase json.RawMessage `json:"option_values_at_purchase"`
	Quantity               int             `json:"quantity"`
	ID                     uuid.UUID       `json:"id"`
	OrgID                  uuid.UUID       `json:"org_id"`
	TenantID               uuid.UUID       `json:"tenant_id"`
	OrderID                uuid.UUID       `json:"order_id"`
}
