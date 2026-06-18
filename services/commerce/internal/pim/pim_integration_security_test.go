package pim_test

import (
	"context"
	"testing"

	"commerce_modules/internal/models"
	"commerce_modules/internal/pim"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateVariantCrossTenantProduct(t *testing.T) {
	ctx := context.Background()
	repo := pim.NewRepository()
	svc := pim.NewService(testPool, repo, nil)

	// Org A
	orgA := uuid.New()
	tenantA := uuid.New()
	productA := uuid.New()
	prodA := &models.Product{
		ID:       productA,
		OrgID:    orgA,
		TenantID: tenantA,
		Title:    "Product A",
		Status:   "ACTIVE",
	}
	if err := svc.CreateProduct(ctx, prodA); err != nil {
		t.Fatalf("Failed to create product A: %v", err)
	}

	// Org B tries to create a variant under Product A
	orgB := uuid.New()
	tenantB := uuid.New()
	sku := "SKU-B"

	variantB := &models.ProductVariant{
		ID:        uuid.New(),
		OrgID:     orgB,
		TenantID:  tenantB,
		ProductID: productA, // referencing Product A from Org A
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(10.0),
	}

	err := svc.CreateVariant(ctx, variantB)
	if err == nil {
		t.Errorf("SECURITY BUG: Successfully created a variant for a product belonging to another tenant!")
	}
}
