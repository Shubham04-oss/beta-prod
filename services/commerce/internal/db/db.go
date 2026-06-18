package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

func NewDB(ctx context.Context, configStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(configStr)
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Dynamic schema updates for tenant-scoped constraints on product_variants
	queries := []string{
		`ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS product_variants_sku_key;`,
		`ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS product_variants_barcode_key;`,
		`ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS product_variants_tenant_id_sku_key;`,
		`ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS product_variants_tenant_id_barcode_key;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_sku_active ON product_variants (tenant_id, sku) WHERE deleted_at IS NULL;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_barcode_active ON product_variants (tenant_id, barcode) WHERE deleted_at IS NULL;`,
	}
	for _, q := range queries {
		_, err := pool.Exec(ctx, q)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
				continue // ignore errors as table might not exist yet
			}
			return nil, err
		}
	}

	return pool, nil
}
