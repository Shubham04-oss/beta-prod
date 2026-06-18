package pim_test

import (
	"context"
	"testing"

	"commerce_modules/internal/models"
	"commerce_modules/internal/pim"

	"github.com/google/uuid"
)

func TestProductNullEmbedding(t *testing.T) {
	ctx := context.Background()
	repo := pim.NewRepository()
	svc := pim.NewService(testPool, repo, nil)

	orgID := uuid.New()
	tenantID := uuid.New()
	productID := uuid.New()

	prod := &models.Product{
		ID:       productID,
		OrgID:    orgID,
		TenantID: tenantID,
		Title:    "Test Product Null Embedding",
		Status:   "ACTIVE",
		// Embedding left as nil
	}

	if err := svc.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	fetched, err := svc.GetProduct(ctx, orgID, tenantID, productID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	if fetched.Embedding != nil {
		t.Fatalf("Expected nil embedding, got %v", fetched.Embedding)
	}
}
