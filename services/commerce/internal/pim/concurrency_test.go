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

func TestCreateVariantOnConcurrentProductDeletion(t *testing.T) {
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
		Title:    "Race Product",
		Status:   "ACTIVE",
	}

	if err := svc.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// We will attempt to run CreateVariant and DeleteProduct concurrently.
	// Since reproducing race condition without locks or sleeps might be flaky, we run it in a loop.
	// Wait, we can't loop easily with the same productID because it gets deleted.
	// We will create 100 products and race them.

	raceFound := false
	for i := 0; i < 50; i++ {
		pID := uuid.New()
		err := svc.CreateProduct(ctx, &models.Product{
			ID:       pID,
			OrgID:    orgID,
			TenantID: tenantID,
			Title:    "Race Product",
			Status:   "ACTIVE",
		})
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}

		vID := uuid.New()
		sku := "SKU-" + uuid.New().String()

		// use channels to synchronize start
		startCh := make(chan struct{})
		var wgLoop sync.WaitGroup
		wgLoop.Add(2)

		go func(productID uuid.UUID) {
			defer wgLoop.Done()
			<-startCh
			_ = svc.DeleteProduct(ctx, orgID, tenantID, productID)
		}(pID)

		go func(productID, variantID uuid.UUID, sku string) {
			defer wgLoop.Done()
			<-startCh
			_ = svc.CreateVariant(ctx, &models.ProductVariant{
				ID:        variantID,
				OrgID:     orgID,
				TenantID:  tenantID,
				ProductID: productID,
				SKU:       &sku,
				Currency:  "USD",
				Price:     decimal.NewFromFloat(10.0),
			})
		}(pID, vID, sku)

		close(startCh)
		wgLoop.Wait()

		// Check if the variant exists and is NOT soft-deleted, but the product IS soft-deleted
		_, errProd := svc.GetProduct(ctx, orgID, tenantID, pID)
		v, errVar := svc.GetVariant(ctx, orgID, tenantID, vID)

		if errProd != nil && errProd == pim.ErrNotFound {
			// product is soft deleted
			if errVar == nil && v != nil && v.DeletedAt == nil {
				// Variant exists and is active!
				raceFound = true
				break
			}
		}
	}

	if raceFound {
		t.Logf("Successfully reproduced race condition: Variant created on a deleted product")
	} else {
		// Might not always hit, but we log it
		t.Logf("Did not hit race condition this time")
	}
}
