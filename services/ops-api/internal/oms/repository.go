package oms

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries db.Querier
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: db.New(pool),
	}
}

// WithTx wraps operations in a transaction
func (r *Repository) WithTx(ctx context.Context, fn func(pgx.Tx, db.Querier) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Automatically enforce RLS for all transactions running through the repository
	tenantID, err := authcontext.GetTenantID(ctx)
	if err == nil && tenantID != "" {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", tenantID); err != nil {
			return err
		}
	}
	orgID, err := authcontext.GetOrgID(ctx)
	if err == nil && orgID != "" {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID); err != nil {
			return err
		}
	}

	qtx := db.New(tx)
	if err := fn(tx, qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
