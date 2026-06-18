package pim_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"commerce_modules/internal/models"
	"commerce_modules/internal/pim"
)

func TestCreateVariantDeleteProductRace(t *testing.T) {
	ctx := context.Background()
	repo := pim.NewRepository()
	inventory := &mockInventoryIntegration{}
	svc := pim.NewService(testPool, repo, inventory)

	orgID := uuid.New()
	tenantID := uuid.New()
	productID := uuid.New()

	prod := &models.Product{
		ID:       productID,
		OrgID:    orgID,
		TenantID: tenantID,
		Title:    "Race Product",
		Status:   "ACTIVE",
	}

	if err := svc.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		sku := "RACE-SKU"
		variant := &models.ProductVariant{
			ID:        uuid.New(),
			OrgID:     orgID,
			TenantID:  tenantID,
			ProductID: productID,
			SKU:       &sku,
			Currency:  "USD",
			Price:     decimal.NewFromFloat(10.0),
		}
		errChan <- svc.CreateVariant(ctx, variant)
	}()

	go func() {
		defer wg.Done()
		errChan <- svc.DeleteProduct(ctx, orgID, tenantID, productID)
	}()

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, err := range errs {
		t.Logf("Concurrent op error: %v", err)
	}

	variants, err := repo.ListVariants(ctx, testPool, orgID, tenantID, productID)
	if err != nil {
		t.Fatalf("Failed to list variants: %v", err)
	}
	if len(variants) > 0 {
		t.Errorf("Expected 0 variants due to cascade delete or CreateVariant failure, got %d", len(variants))
	}
}
