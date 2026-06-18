package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/synq/ops-api/internal/pim"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, topic, eventType string, data interface{}) error {
	return nil
}
func (m *MockPublisher) Close() error { return nil }

func main() {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	if err != nil {
		fmt.Printf("Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var orgID, tenantID uuid.UUID
	dbpool.QueryRow(ctx, "SELECT id FROM organizations LIMIT 1").Scan(&orgID)
	dbpool.QueryRow(ctx, "SELECT id FROM tenants WHERE org_id = $1 LIMIT 1", orgID).Scan(&tenantID)

	authCtx := context.WithValue(ctx, authcontext.TenantIDKey, tenantID.String())
	authCtx = context.WithValue(authCtx, authcontext.OrgIDKey, orgID.String())

	pimSvc := pim.NewService(dbpool, &MockPublisher{})

	prodID := uuid.New()
	newProduct, err := pimSvc.CreateProduct(authCtx, db.CreateProductParams{
		ID:       pgtype.UUID{Bytes: prodID, Valid: true},
		OrgID:    pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID: pgtype.UUID{Bytes: tenantID, Valid: true},
		Title:    "E2E Test Product",
		Status:   pgtype.Text{String: "ACTIVE", Valid: true},
	})
	fmt.Println("Create Product err: ", err)

	variantID := uuid.New()
	_, err = pimSvc.CreateVariant(authCtx, db.CreateProductVariantParams{
		ID:        pgtype.UUID{Bytes: variantID, Valid: true},
		OrgID:     pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:  pgtype.UUID{Bytes: tenantID, Valid: true},
		ProductID: newProduct.ID,
		Sku:       pgtype.Text{String: "E2E-SKU-001", Valid: true},
	})
	fmt.Println("Create Variant err: ", err)
}
