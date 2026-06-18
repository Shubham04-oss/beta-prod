package pim

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pgvector/pgvector-go"

	"commerce_modules/internal/models"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrDuplicateSKU = errors.New("duplicate sku or barcode")
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) CreateProduct(ctx context.Context, db DBTX, product *models.Product) error {
	query := `
		INSERT INTO products (
			id, org_id, tenant_id, created_at, updated_at,
			created_by, updated_by, title, description, status,
			options, metadata, embedding
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13
		)
	`
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	_, err := db.Exec(ctx, query,
		product.ID, product.OrgID, product.TenantID, product.CreatedAt, product.UpdatedAt,
		product.CreatedBy, product.UpdatedBy, product.Title, product.Description, product.Status,
		product.Options, product.Metadata, product.Embedding,
	)
	return err
}

func (r *Repository) GetProduct(ctx context.Context, db DBTX, orgID, tenantID, productID uuid.UUID) (*models.Product, error) {
	query := `
		SELECT
			id, org_id, tenant_id, created_at, updated_at, deleted_at,
			created_by, updated_by, title, description, status,
			options, metadata, embedding
		FROM products
		WHERE org_id = $1 AND tenant_id = $2 AND id = $3 AND deleted_at IS NULL
	`

	var p models.Product
	err := db.QueryRow(ctx, query, orgID, tenantID, productID).Scan(
		&p.ID, &p.OrgID, &p.TenantID, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		&p.CreatedBy, &p.UpdatedBy, &p.Title, &p.Description, &p.Status,
		&p.Options, &p.Metadata, &p.Embedding,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *Repository) UpdateProduct(ctx context.Context, db DBTX, product *models.Product) error {
	query := `
		UPDATE products SET
			title = $1, description = $2, status = $3, options = $4,
			metadata = $5, embedding = $6, updated_at = $7, updated_by = $8
		WHERE org_id = $9 AND tenant_id = $10 AND id = $11 AND deleted_at IS NULL
	`
	product.UpdatedAt = time.Now()

	cmd, err := db.Exec(ctx, query,
		product.Title, product.Description, product.Status, product.Options,
		product.Metadata, product.Embedding, product.UpdatedAt, product.UpdatedBy,
		product.OrgID, product.TenantID, product.ID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteProduct(ctx context.Context, db DBTX, orgID, tenantID, productID uuid.UUID) error {
	query := `
		UPDATE products
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE org_id = $1 AND tenant_id = $2 AND id = $3 AND deleted_at IS NULL
	`
	cmd, err := db.Exec(ctx, query, orgID, tenantID, productID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Cascading soft delete for variants
	queryVariants := `
		UPDATE product_variants
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE org_id = $1 AND tenant_id = $2 AND product_id = $3 AND deleted_at IS NULL
	`
	_, err = db.Exec(ctx, queryVariants, orgID, tenantID, productID)
	return err
}

func (r *Repository) CreateVariant(ctx context.Context, db DBTX, variant *models.ProductVariant) error {
	query := `
		INSERT INTO product_variants (
			id, org_id, tenant_id, product_id, created_at, updated_at,
			created_by, updated_by, sku, barcode, currency, price,
			option_values, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14
		)
	`
	now := time.Now()
	variant.CreatedAt = now
	variant.UpdatedAt = now

	_, err := db.Exec(ctx, query,
		variant.ID, variant.OrgID, variant.TenantID, variant.ProductID, variant.CreatedAt, variant.UpdatedAt,
		variant.CreatedBy, variant.UpdatedBy, variant.SKU, variant.Barcode, variant.Currency, variant.Price,
		variant.OptionValues, variant.Metadata,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicateSKU
	}

	return err
}

func (r *Repository) GetVariant(ctx context.Context, db DBTX, orgID, tenantID, variantID uuid.UUID) (*models.ProductVariant, error) {
	query := `
		SELECT
			id, org_id, tenant_id, product_id, created_at, updated_at, deleted_at,
			created_by, updated_by, sku, barcode, currency, price,
			option_values, metadata
		FROM product_variants
		WHERE org_id = $1 AND tenant_id = $2 AND id = $3 AND deleted_at IS NULL
	`

	var v models.ProductVariant
	err := db.QueryRow(ctx, query, orgID, tenantID, variantID).Scan(
		&v.ID, &v.OrgID, &v.TenantID, &v.ProductID, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
		&v.CreatedBy, &v.UpdatedBy, &v.SKU, &v.Barcode, &v.Currency, &v.Price,
		&v.OptionValues, &v.Metadata,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &v, err
}

func (r *Repository) ListVariants(ctx context.Context, db DBTX, orgID, tenantID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	query := `
		SELECT
			id, org_id, tenant_id, product_id, created_at, updated_at, deleted_at,
			created_by, updated_by, sku, barcode, currency, price,
			option_values, metadata
		FROM product_variants
		WHERE org_id = $1 AND tenant_id = $2 AND product_id = $3 AND deleted_at IS NULL
	`

	rows, err := db.Query(ctx, query, orgID, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []*models.ProductVariant
	for rows.Next() {
		var v models.ProductVariant
		if err := rows.Scan(
			&v.ID, &v.OrgID, &v.TenantID, &v.ProductID, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
			&v.CreatedBy, &v.UpdatedBy, &v.SKU, &v.Barcode, &v.Currency, &v.Price,
			&v.OptionValues, &v.Metadata,
		); err != nil {
			return nil, err
		}
		variants = append(variants, &v)
	}
	return variants, rows.Err()
}

func (r *Repository) UpdateVariant(ctx context.Context, db DBTX, variant *models.ProductVariant) error {
	query := `
		UPDATE product_variants SET
			sku = $1, barcode = $2, currency = $3, price = $4,
			option_values = $5, metadata = $6, updated_at = $7, updated_by = $8
		WHERE org_id = $9 AND tenant_id = $10 AND id = $11 AND deleted_at IS NULL
	`
	variant.UpdatedAt = time.Now()

	cmd, err := db.Exec(ctx, query,
		variant.SKU, variant.Barcode, variant.Currency, variant.Price,
		variant.OptionValues, variant.Metadata, variant.UpdatedAt, variant.UpdatedBy,
		variant.OrgID, variant.TenantID, variant.ID,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicateSKU
	}

	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteVariant(ctx context.Context, db DBTX, orgID, tenantID, variantID uuid.UUID) error {
	query := `
		UPDATE product_variants
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE org_id = $1 AND tenant_id = $2 AND id = $3 AND deleted_at IS NULL
	`
	cmd, err := db.Exec(ctx, query, orgID, tenantID, variantID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) SearchProducts(ctx context.Context, db DBTX, orgID, tenantID uuid.UUID, embedding pgvector.Vector, limit int) ([]*models.Product, error) {
	query := `
		SELECT
			id, org_id, tenant_id, created_at, updated_at, deleted_at,
			created_by, updated_by, title, description, status,
			options, metadata, embedding
		FROM products
		WHERE org_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY embedding <=> $3
		LIMIT $4
	`

	rows, err := db.Query(ctx, query, orgID, tenantID, embedding, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.OrgID, &p.TenantID, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
			&p.CreatedBy, &p.UpdatedBy, &p.Title, &p.Description, &p.Status,
			&p.Options, &p.Metadata, &p.Embedding,
		); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, rows.Err()
}

func (r *Repository) LockProduct(ctx context.Context, db DBTX, orgID, tenantID, productID uuid.UUID, lockClause string) error {
	query := `SELECT id FROM products WHERE org_id = $1 AND tenant_id = $2 AND id = $3 AND deleted_at IS NULL ` + lockClause
	var id uuid.UUID
	err := db.QueryRow(ctx, query, orgID, tenantID, productID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
