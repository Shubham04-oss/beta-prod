package pim_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"commerce_modules/internal/models"
	"commerce_modules/internal/pim"
)

type mockDBTX struct {
	execFunc     func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	queryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockDBTX) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, arguments...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return nil
}

func TestRepository_CreateProduct(t *testing.T) {
	repo := pim.NewRepository()
	product := &models.Product{
		ID:       uuid.New(),
		OrgID:    uuid.New(),
		TenantID: uuid.New(),
		Title:    "Test",
	}

	mockDB := &mockDBTX{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			if len(arguments) != 13 {
				t.Errorf("Expected 13 arguments, got %d", len(arguments))
			}
			return pgconn.CommandTag{}, nil
		},
	}

	err := repo.CreateProduct(context.Background(), mockDB, product)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
