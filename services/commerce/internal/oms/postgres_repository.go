package oms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"commerce_modules/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateOrderWithLineItems(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO orders (id, org_id, tenant_id, customer_id, currency, status, total_price, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	now := time.Now()
	order.TenantID = tenantID
	order.OrgID = orgID
	order.CreatedAt = now
	order.UpdatedAt = now

	_, err = tx.Exec(ctx, orderQuery, order.ID, order.OrgID, order.TenantID, order.CustomerID, order.Currency, order.Status, order.TotalPrice, order.Metadata, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range items {
		item := &items[i]
		item.CreatedAt = now
		item.OrderID = order.ID
		item.OrgID = order.OrgID
		item.TenantID = order.TenantID

		itemQuery := `
			INSERT INTO order_line_items (id, org_id, tenant_id, order_id, variant_id, price_at_purchase, option_values_at_purchase, quantity, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.Exec(ctx, itemQuery, item.ID, item.OrgID, item.TenantID, item.OrderID, item.VariantID, item.PriceAtPurchase, item.OptionValuesAtPurchase, item.Quantity, item.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) UpdateOrderStatus(ctx context.Context, tenantID, orgID, orderID uuid.UUID, currentStatus, newStatus models.OrderStatus) error {
	query := `
		UPDATE orders 
		SET status = $1, updated_at = $2
		WHERE id = $3 AND org_id = $4 AND tenant_id = $5 AND status = $6 AND deleted_at IS NULL
	`
	now := time.Now()
	tag, err := r.db.Exec(ctx, query, newStatus, now, orderID, orgID, tenantID, currentStatus)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("order not found or in invalid state")
	}
	return nil
}

func (r *PostgresRepository) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	query := `
		SELECT id, org_id, tenant_id, customer_id, currency, status, total_price, metadata, created_at, updated_at, deleted_at
		FROM orders
		WHERE id = $1 AND org_id = $2 AND tenant_id = $3 AND deleted_at IS NULL
	`
	row := r.db.QueryRow(ctx, query, orderID, orgID, tenantID)

	var o models.Order
	err := row.Scan(&o.ID, &o.OrgID, &o.TenantID, &o.CustomerID, &o.Currency, &o.Status, &o.TotalPrice, &o.Metadata, &o.CreatedAt, &o.UpdatedAt, &o.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, err
	}
	return &o, nil
}

func (r *PostgresRepository) GetOrderLineItems(ctx context.Context, tenantID, orgID, orderID uuid.UUID) ([]models.OrderLineItem, error) {
	query := `
		SELECT id, org_id, tenant_id, order_id, variant_id, price_at_purchase, option_values_at_purchase, quantity, created_at
		FROM order_line_items
		WHERE order_id = $1 AND org_id = $2 AND tenant_id = $3
	`
	rows, err := r.db.Query(ctx, query, orderID, orgID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderLineItem
	for rows.Next() {
		var item models.OrderLineItem
		err := rows.Scan(&item.ID, &item.OrgID, &item.TenantID, &item.OrderID, &item.VariantID, &item.PriceAtPurchase, &item.OptionValuesAtPurchase, &item.Quantity, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	query := `
		INSERT INTO customers (id, org_id, tenant_id, first_name, last_name, email, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	now := time.Now()
	customer.TenantID = tenantID
	customer.OrgID = orgID
	customer.CreatedAt = now
	customer.UpdatedAt = now

	_, err := r.db.Exec(ctx, query, customer.ID, customer.OrgID, customer.TenantID, customer.FirstName, customer.LastName, customer.Email, customer.Metadata, customer.CreatedAt, customer.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) (*models.Customer, error) {
	query := `
		SELECT id, org_id, tenant_id, first_name, last_name, email, metadata, created_at, updated_at, deleted_at
		FROM customers
		WHERE id = $1 AND org_id = $2 AND tenant_id = $3 AND deleted_at IS NULL
	`
	row := r.db.QueryRow(ctx, query, customerID, orgID, tenantID)

	var c models.Customer
	err := row.Scan(&c.ID, &c.OrgID, &c.TenantID, &c.FirstName, &c.LastName, &c.Email, &c.Metadata, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("customer not found: %w", err)
		}
		return nil, err
	}
	return &c, nil
}

func (r *PostgresRepository) UpdateCustomer(ctx context.Context, tenantID, orgID uuid.UUID, customer *models.Customer) error {
	query := `
		UPDATE customers
		SET first_name = $1, last_name = $2, email = $3, metadata = $4, updated_at = $5
		WHERE id = $6 AND org_id = $7 AND tenant_id = $8 AND deleted_at IS NULL
	`
	customer.TenantID = tenantID
	customer.OrgID = orgID
	customer.UpdatedAt = time.Now()
	tag, err := r.db.Exec(ctx, query, customer.FirstName, customer.LastName, customer.Email, customer.Metadata, customer.UpdatedAt, customer.ID, customer.OrgID, customer.TenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("customer not found")
	}
	return nil
}

func (r *PostgresRepository) DeleteCustomer(ctx context.Context, tenantID, orgID, customerID uuid.UUID) error {
	query := `
		UPDATE customers
		SET deleted_at = $1
		WHERE id = $2 AND org_id = $3 AND tenant_id = $4 AND deleted_at IS NULL
	`
	now := time.Now()
	tag, err := r.db.Exec(ctx, query, now, customerID, orgID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("customer not found")
	}
	return nil
}
