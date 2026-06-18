package pim_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"commerce_modules/internal/models"
	"commerce_modules/internal/pim"
)

type mockInventoryIntegration struct {
	failInitialize error
	failCascade    error
	initCalled     bool
	cascadeCalled  bool
}

func (m *mockInventoryIntegration) InitializeInventory(ctx context.Context, tx pgx.Tx, variantID uuid.UUID) error {
	m.initCalled = true
	return m.failInitialize
}

func (m *mockInventoryIntegration) CascadeVariantDeletion(ctx context.Context, tx pgx.Tx, variantID uuid.UUID) error {
	m.cascadeCalled = true
	return m.failCascade
}

var testPool *pgxpool.Pool

func getFreePort() uint32 {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return uint32(l.Addr().(*net.TCPAddr).Port)
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	port := getFreePort()

	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "pim_test_")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	config := embeddedpostgres.DefaultConfig().
		Port(port).
		DataPath(filepath.Join(tempDir, "data")).
		RuntimePath(filepath.Join(tempDir, "runtime"))

	postgresContainer := embeddedpostgres.NewDatabase(config)
	if err := postgresContainer.Start(); err != nil {
		panic(err)
	}

	defer func() {
		if err := postgresContainer.Stop(); err != nil {
			panic(err)
		}
	}()

	connStr := fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic(err)
	}
	testPool = pool

	// pgvector might not be available
	_, err = testPool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector;")
	if err != nil {
		// Just skip vector column if not possible
		_, _ = testPool.Exec(ctx, `
			CREATE TABLE products (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				deleted_at TIMESTAMP,
				created_by UUID,
				updated_by UUID,
				title TEXT NOT NULL,
				description TEXT,
				status TEXT NOT NULL,
				options JSONB,
				metadata JSONB,
				embedding text -- dummy
			);
		`)
	} else {
		_, _ = testPool.Exec(ctx, `
			CREATE TABLE products (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				deleted_at TIMESTAMP,
				created_by UUID,
				updated_by UUID,
				title TEXT NOT NULL,
				description TEXT,
				status TEXT NOT NULL,
				options JSONB,
				metadata JSONB,
				embedding vector(3)
			);
		`)
	}

	_, err = testPool.Exec(ctx, `
		CREATE TABLE product_variants (
			id UUID PRIMARY KEY,
			org_id UUID NOT NULL,
			tenant_id UUID NOT NULL,
			product_id UUID NOT NULL REFERENCES products(id),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP,
			created_by UUID,
			updated_by UUID,
			sku TEXT,
			barcode TEXT,
			currency TEXT NOT NULL,
			price NUMERIC NOT NULL,
			option_values JSONB,
			metadata JSONB
		);
	`)
	if err != nil {
		panic(err)
	}

	_, err = testPool.Exec(ctx, `
		CREATE UNIQUE INDEX idx_tenant_sku_active ON product_variants (tenant_id, sku) WHERE deleted_at IS NULL;
		CREATE UNIQUE INDEX idx_tenant_barcode_active ON product_variants (tenant_id, barcode) WHERE deleted_at IS NULL;
	`)
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestDuplicateSKU(t *testing.T) {
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
		Title:    "Test Product",
		Status:   "ACTIVE",
	}

	if err := svc.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	sku := "SKU-123"
	variant1 := &models.ProductVariant{
		ID:        uuid.New(),
		OrgID:     orgID,
		TenantID:  tenantID,
		ProductID: productID,
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(10.0),
	}

	if err := svc.CreateVariant(ctx, variant1); err != nil {
		t.Fatalf("Failed to create variant 1: %v", err)
	}

	variant2 := &models.ProductVariant{
		ID:        uuid.New(),
		OrgID:     orgID,
		TenantID:  tenantID,
		ProductID: productID,
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(20.0),
	}

	err := svc.CreateVariant(ctx, variant2)
	if err == nil {
		t.Fatalf("Expected error for duplicate SKU, got nil")
	}
	if !errors.Is(err, pim.ErrDuplicateSKU) {
		t.Fatalf("Expected ErrDuplicateSKU, got: %v", err)
	}
}

func TestTransactionalRollback(t *testing.T) {
	ctx := context.Background()
	repo := pim.NewRepository()

	inventory := &mockInventoryIntegration{
		failInitialize: errors.New("inventory system down"),
	}

	svc := pim.NewService(testPool, repo, inventory)

	orgID := uuid.New()
	tenantID := uuid.New()
	productID := uuid.New()

	prod := &models.Product{
		ID:       productID,
		OrgID:    orgID,
		TenantID: tenantID,
		Title:    "Test Product Rollback",
		Status:   "ACTIVE",
	}

	if err := svc.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	variantID := uuid.New()
	sku := "SKU-ROLLBACK"
	variant := &models.ProductVariant{
		ID:        variantID,
		OrgID:     orgID,
		TenantID:  tenantID,
		ProductID: productID,
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(10.0),
	}

	err := svc.CreateVariant(ctx, variant)
	if err == nil {
		t.Fatalf("Expected error due to inventory failure")
	}

	// Verify that the variant was NOT created in DB (rolled back)
	_, err = repo.GetVariant(ctx, testPool, orgID, tenantID, variantID)
	if err == nil {
		t.Fatalf("Expected variant to be not found due to rollback, but it exists")
	}
	if !errors.Is(err, pim.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got: %v", err)
	}
}

func TestDeleteProductCascadesVariants(t *testing.T) {
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
		Title:    "Test Product Delete",
		Status:   "ACTIVE",
	}

	if err := svc.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	variantID := uuid.New()
	sku := "SKU-DELETE"
	variant := &models.ProductVariant{
		ID:        variantID,
		OrgID:     orgID,
		TenantID:  tenantID,
		ProductID: productID,
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(10.0),
	}

	if err := svc.CreateVariant(ctx, variant); err != nil {
		t.Fatalf("Failed to create variant: %v", err)
	}

	// Now delete product
	if err := svc.DeleteProduct(ctx, orgID, tenantID, productID); err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	// Verify inventory cascade was called
	if !inventory.cascadeCalled {
		t.Fatalf("Expected inventory cascade to be called")
	}

	// Verify product deleted
	_, err := repo.GetProduct(ctx, testPool, orgID, tenantID, productID)
	if !errors.Is(err, pim.ErrNotFound) {
		t.Fatalf("Expected product ErrNotFound, got %v", err)
	}

	// Verify variant deleted
	_, err = repo.GetVariant(ctx, testPool, orgID, tenantID, variantID)
	if !errors.Is(err, pim.ErrNotFound) {
		t.Fatalf("Expected variant ErrNotFound, got %v", err)
	}
}

func TestCreateVariantOnDeletedProduct(t *testing.T) {
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
		Title:    "Test Product Delete Then Variant",
		Status:   "ACTIVE",
	}

	if err := svc.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Delete product
	if err := svc.DeleteProduct(ctx, orgID, tenantID, productID); err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	sku := "SKU-SHOULD-FAIL"
	variant := &models.ProductVariant{
		ID:        uuid.New(),
		OrgID:     orgID,
		TenantID:  tenantID,
		ProductID: productID,
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(10.0),
	}

	err := svc.CreateVariant(ctx, variant)
	if err == nil {
		t.Errorf("Expected error when creating variant for a soft-deleted product, but got none!")
	}
}

func TestCrossTenantDuplicateSKU(t *testing.T) {
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
		Title:    "Test Product A",
		Status:   "ACTIVE",
	}
	_ = svc.CreateProduct(ctx, prodA)

	sku := "SHARED-SKU"
	variantA := &models.ProductVariant{
		ID:        uuid.New(),
		OrgID:     orgA,
		TenantID:  tenantA,
		ProductID: productA,
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(10.0),
	}
	if err := svc.CreateVariant(ctx, variantA); err != nil {
		t.Fatalf("Failed to create variant A: %v", err)
	}

	// Org B
	orgB := uuid.New()
	tenantB := uuid.New()
	productB := uuid.New()
	prodB := &models.Product{
		ID:       productB,
		OrgID:    orgB,
		TenantID: tenantB,
		Title:    "Test Product B",
		Status:   "ACTIVE",
	}
	_ = svc.CreateProduct(ctx, prodB)

	variantB := &models.ProductVariant{
		ID:        uuid.New(),
		OrgID:     orgB,
		TenantID:  tenantB,
		ProductID: productB,
		SKU:       &sku,
		Currency:  "USD",
		Price:     decimal.NewFromFloat(10.0),
	}
	err := svc.CreateVariant(ctx, variantB)
	if err != nil {
		t.Errorf("Expected to be able to create variant with same SKU in different org, but got error: %v", err)
	}
}
