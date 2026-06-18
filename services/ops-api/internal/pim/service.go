package pim

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/synq/ops-api/internal/telemetry"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
	"github.com/synq/pkg/events"
)

// Service defines the boundary for Product Information Management.
type Service interface {
	CreateProduct(ctx context.Context, params db.CreateProductParams) (db.Product, error)
	CreateVariant(ctx context.Context, params db.CreateProductVariantParams) (db.ProductVariant, error)
	CreateMedia(ctx context.Context, params db.CreateProductMediaParams) (db.ProductMedium, error)

	GetProduct(ctx context.Context, params db.GetProductParams) (db.Product, error)
	UpdateProduct(ctx context.Context, params db.UpdateProductParams) (db.Product, error)
	DeleteProduct(ctx context.Context, params db.DeleteProductParams) error

	GetDashboardStats(ctx context.Context) (PIMDashboardStats, error)
	AdjustInventory(ctx context.Context, params db.CreateInventoryLedgerEntryParams) (db.InventoryLedger, error)

	CreateBrand(ctx context.Context, params db.CreateBrandParams) (db.Brand, error)
	GetBrands(ctx context.Context, tenantID pgtype.UUID) ([]db.Brand, error)

	CreateCategory(ctx context.Context, params db.CreateCategoryParams) (db.Category, error)
	GetCategories(ctx context.Context, tenantID pgtype.UUID) ([]db.Category, error)

	CreateAttribute(ctx context.Context, params db.CreateAttributeParams) (db.Attribute, error)
	GetAttributes(ctx context.Context, tenantID pgtype.UUID) ([]db.Attribute, error)

	CreateProductType(ctx context.Context, params db.CreateProductTypeParams) (db.ProductType, error)
	GetProductTypes(ctx context.Context, tenantID pgtype.UUID) ([]db.ProductType, error)

	GetProductMedia(ctx context.Context, params db.GetProductMediaParams) ([]db.ProductMedium, error)
	GetAuditEvents(ctx context.Context, tenantID pgtype.UUID) ([]db.AuditEvent, error)
	GetValidationIssues(ctx context.Context, tenantID pgtype.UUID) ([]db.ValidationIssue, error)

	CreateAttributeGroup(ctx context.Context, params db.CreateAttributeGroupParams) (db.AttributeGroup, error)
	GetAttributeGroups(ctx context.Context, tenantID pgtype.UUID) ([]db.AttributeGroup, error)

	CreateBulkJob(ctx context.Context, params db.CreateBulkJobParams) (db.BulkJob, error)
	GetBulkJobs(ctx context.Context, tenantID pgtype.UUID, jobType string) ([]db.BulkJob, error)
}

type PIMDashboardStats struct {
	TotalProducts       int64                          `json:"totalProducts"`
	LowStockVariants    int32                          `json:"lowStockVariants"`
	OutOfStockVariants  int32                          `json:"outOfStockVariants"`
	TotalInventoryValue float64                        `json:"totalInventoryValue"`
	TopLowStock         []db.GetTopLowStockProductsRow `json:"topLowStock"`
}

type service struct {
	pool      *pgxpool.Pool
	publisher events.Publisher
}

// NewService creates a new PIM service.
func NewService(pool *pgxpool.Pool, publisher events.Publisher) Service {
	return &service{
		pool:      pool,
		publisher: publisher,
	}
}

// enforceRLS wraps a transaction with the strict 4-ID RLS setup.
func (s *service) enforceRLS(ctx context.Context, tx db.DBTX) error {
	tenantID, err := authcontext.GetTenantID(ctx)
	if err != nil {
		return fmt.Errorf("unauthorized: missing tenant_id")
	}
	orgID, _ := authcontext.GetOrgID(ctx) // Optional in some contexts, but good to have

	// Set Postgres session variables for RLS
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", tenantID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID)
	return err
}

func (s *service) CreateProduct(ctx context.Context, params db.CreateProductParams) (db.Product, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Product{}, err
	}
	defer tx.Rollback(ctx)

	// Enforce strict multi-tenancy before executing any queries
	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.Product{}, err
	}

	queries := db.New(tx)

	// Create the product in the database
	productRow, err := queries.CreateProduct(ctx, params)
	if err != nil {
		return db.Product{}, fmt.Errorf("failed to create product: %w", err)
	}

	product := db.Product{
		ID:                   productRow.ID,
		OrgID:                productRow.OrgID,
		TenantID:             productRow.TenantID,
		Title:                productRow.Title,
		Description:          productRow.Description,
		Status:               productRow.Status,
		ProductTypeID:        productRow.ProductTypeID,
		CreatedBy:            productRow.CreatedBy,
		UpdatedBy:            productRow.UpdatedBy,
		CreatedAt:            productRow.CreatedAt,
		UpdatedAt:            productRow.UpdatedAt,
		DeletedAt:            productRow.DeletedAt,
		ShortDescription:     productRow.ShortDescription,
		LongDescription:      productRow.LongDescription,
		Category:             productRow.Category,
		Brand:                productRow.Brand,
		Tags:                 productRow.Tags,
		LaunchDate:           productRow.LaunchDate,
		DiscontinueDate:      productRow.DiscontinueDate,
		WarrantyPeriod:       productRow.WarrantyPeriod,
		IsTaxable:            productRow.IsTaxable,
		IsReturnable:         productRow.IsReturnable,
		RequiresSerialNumber: productRow.RequiresSerialNumber,
		SeoTitle:             productRow.SeoTitle,
		SeoDescription:       productRow.SeoDescription,
		SeoKeywords:          productRow.SeoKeywords,
		DataQualityScore:     productRow.DataQualityScore,
	}

	// Commit transaction before publishing event
	if err := tx.Commit(ctx); err != nil {
		return db.Product{}, err
	}

	// Record Prometheus Metric
	if tenantIDStr, err := authcontext.GetTenantID(ctx); err == nil {
		telemetry.PIMProductsCreatedTotal.WithLabelValues(tenantIDStr).Inc()
	}

	// Fire the Domain Event to Pub/Sub
	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.product.created", product)
	if err != nil {
		// Log warning, but product is created
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return product, nil
}

func (s *service) CreateVariant(ctx context.Context, params db.CreateProductVariantParams) (db.ProductVariant, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.ProductVariant{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.ProductVariant{}, err
	}

	queries := db.New(tx)

	variant, err := queries.CreateProductVariant(ctx, params)
	if err != nil {
		return db.ProductVariant{}, fmt.Errorf("failed to create variant: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.ProductVariant{}, err
	}

	// Fire the Domain Event to Pub/Sub
	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.variant.created", variant)
	if err != nil {
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return variant, nil
}

func (s *service) CreateMedia(ctx context.Context, params db.CreateProductMediaParams) (db.ProductMedium, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.ProductMedium{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.ProductMedium{}, err
	}

	queries := db.New(tx)

	media, err := queries.CreateProductMedia(ctx, params)
	if err != nil {
		return db.ProductMedium{}, fmt.Errorf("failed to create media: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.ProductMedium{}, err
	}

	// Fire the Domain Event to Pub/Sub
	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.media.created", media)
	if err != nil {
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return media, nil
}

func (s *service) GetProduct(ctx context.Context, params db.GetProductParams) (db.Product, error) {
	// Not in a transaction just for read, but we could if we wanted to enforce RLS per transaction.
	// Since GetProduct takes tenant_id, it is implicitly safe.
	queries := db.New(s.pool)
	return queries.GetProduct(ctx, params)
}

func (s *service) UpdateProduct(ctx context.Context, params db.UpdateProductParams) (db.Product, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Product{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.Product{}, err
	}

	queries := db.New(tx)

	product, err := queries.UpdateProduct(ctx, params)
	if err != nil {
		return db.Product{}, fmt.Errorf("failed to update product: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Product{}, err
	}

	// Record Prometheus Metric
	if tenantIDStr, err := authcontext.GetTenantID(ctx); err == nil {
		telemetry.PIMProductsUpdatedTotal.WithLabelValues(tenantIDStr).Inc()
	}

	// Fire the Domain Event
	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.product.updated", product)
	if err != nil {
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return product, nil
}

func (s *service) DeleteProduct(ctx context.Context, params db.DeleteProductParams) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return err
	}

	queries := db.New(tx)

	err = queries.DeleteProduct(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Record Prometheus Metric
	if tenantIDStr, err := authcontext.GetTenantID(ctx); err == nil {
		telemetry.PIMProductsDeletedTotal.WithLabelValues(tenantIDStr).Inc()
	}

	// Fire the Domain Event
	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.product.deleted", params)
	if err != nil {
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return nil
}

func (s *service) GetDashboardStats(ctx context.Context) (PIMDashboardStats, error) {
	tenantIDStr, err := authcontext.GetTenantID(ctx)
	if err != nil {
		return PIMDashboardStats{}, fmt.Errorf("unauthorized: missing tenant_id")
	}
	var tenantID pgtype.UUID
	if err := tenantID.Scan(tenantIDStr); err != nil {
		return PIMDashboardStats{}, err
	}

	queries := db.New(s.pool)

	statsRow, err := queries.GetPIMStatsByTenant(ctx, tenantID)
	if err != nil {
		return PIMDashboardStats{}, fmt.Errorf("failed to get stats: %w", err)
	}

	topLow, err := queries.GetTopLowStockProducts(ctx, tenantID)
	if err != nil {
		fmt.Printf("Warning: failed to fetch top low stock products: %v\n", err)
	}

	val, err := statsRow.TotalInventoryValue.Float64Value()
	var totalVal float64
	if err == nil && val.Valid {
		totalVal = val.Float64
	}

	return PIMDashboardStats{
		TotalProducts:       statsRow.TotalProducts,
		LowStockVariants:    statsRow.LowStockVariants,
		OutOfStockVariants:  statsRow.OutOfStockVariants,
		TotalInventoryValue: totalVal,
		TopLowStock:         topLow,
	}, nil
}

func (s *service) AdjustInventory(ctx context.Context, params db.CreateInventoryLedgerEntryParams) (db.InventoryLedger, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.InventoryLedger{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.InventoryLedger{}, err
	}

	queries := db.New(tx)

	ledgerEntry, err := queries.CreateInventoryLedgerEntry(ctx, params)
	if err != nil {
		return db.InventoryLedger{}, fmt.Errorf("failed to create inventory ledger entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.InventoryLedger{}, err
	}

	// Fire the Domain Event to Pub/Sub
	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.inventory.adjusted", ledgerEntry)
	if err != nil {
		fmt.Printf("Warning: failed to publish inventory adjusted event: %v\n", err)
	}

	return ledgerEntry, nil
}

func (s *service) CreateBrand(ctx context.Context, params db.CreateBrandParams) (db.Brand, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Brand{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.Brand{}, err
	}

	queries := db.New(tx)

	brand, err := queries.CreateBrand(ctx, params)
	if err != nil {
		return db.Brand{}, fmt.Errorf("failed to create brand: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Brand{}, err
	}

	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.brand.created", brand)
	if err != nil {
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return brand, nil
}

func (s *service) GetBrands(ctx context.Context, tenantID pgtype.UUID) ([]db.Brand, error) {
	queries := db.New(s.pool)
	return queries.GetBrands(ctx, tenantID)
}

func (s *service) CreateCategory(ctx context.Context, params db.CreateCategoryParams) (db.Category, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Category{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.Category{}, err
	}

	queries := db.New(tx)

	category, err := queries.CreateCategory(ctx, params)
	if err != nil {
		return db.Category{}, fmt.Errorf("failed to create category: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Category{}, err
	}

	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.category.created", category)
	if err != nil {
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return category, nil
}

func (s *service) GetCategories(ctx context.Context, tenantID pgtype.UUID) ([]db.Category, error) {
	queries := db.New(s.pool)
	return queries.GetCategories(ctx, tenantID)
}

func (s *service) CreateAttribute(ctx context.Context, params db.CreateAttributeParams) (db.Attribute, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Attribute{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.enforceRLS(ctx, tx); err != nil {
		return db.Attribute{}, err
	}

	queries := db.New(tx)

	attribute, err := queries.CreateAttribute(ctx, params)
	if err != nil {
		return db.Attribute{}, fmt.Errorf("failed to create attribute: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Attribute{}, err
	}

	err = s.publisher.Publish(ctx, "pim-events", "synq.pim.attribute.created", attribute)
	if err != nil {
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	return attribute, nil
}

func (s *service) GetAttributes(ctx context.Context, tenantID pgtype.UUID) ([]db.Attribute, error) {
	queries := db.New(s.pool)
	return queries.GetAttributes(ctx, tenantID)
}

func (s *service) CreateProductType(ctx context.Context, params db.CreateProductTypeParams) (db.ProductType, error) {
	queries := db.New(s.pool)
	pt, err := queries.CreateProductType(ctx, params)
	if err != nil {
		return db.ProductType{}, err
	}
	s.publisher.Publish(ctx, "pim-events", "synq.pim.template.created", pt)
	return pt, nil
}

func (s *service) GetProductTypes(ctx context.Context, tenantID pgtype.UUID) ([]db.ProductType, error) {
	queries := db.New(s.pool)
	return queries.GetProductTypes(ctx, tenantID)
}

func (s *service) GetProductMedia(ctx context.Context, params db.GetProductMediaParams) ([]db.ProductMedium, error) {
	queries := db.New(s.pool)
	return queries.GetProductMedia(ctx, params)
}

func (s *service) GetAuditEvents(ctx context.Context, tenantID pgtype.UUID) ([]db.AuditEvent, error) {
	queries := db.New(s.pool)
	return queries.GetAuditEvents(ctx, tenantID)
}

func (s *service) GetValidationIssues(ctx context.Context, tenantID pgtype.UUID) ([]db.ValidationIssue, error) {
	queries := db.New(s.pool)
	return queries.GetValidationIssues(ctx, tenantID)
}

func (s *service) CreateAttributeGroup(ctx context.Context, params db.CreateAttributeGroupParams) (db.AttributeGroup, error) {
	queries := db.New(s.pool)
	ag, err := queries.CreateAttributeGroup(ctx, params)
	if err != nil {
		return db.AttributeGroup{}, err
	}
	s.publisher.Publish(ctx, "pim-events", "synq.pim.attribute_group.created", ag)
	return ag, nil
}

func (s *service) GetAttributeGroups(ctx context.Context, tenantID pgtype.UUID) ([]db.AttributeGroup, error) {
	queries := db.New(s.pool)
	return queries.GetAttributeGroups(ctx, tenantID)
}

func (s *service) CreateBulkJob(ctx context.Context, params db.CreateBulkJobParams) (db.BulkJob, error) {
	queries := db.New(s.pool)
	job, err := queries.CreateBulkJob(ctx, params)
	if err != nil {
		return db.BulkJob{}, err
	}
	// Publish an outbox event so workers pick it up
	s.publisher.Publish(ctx, "pim-events", "synq.pim.bulk_job.started", job)
	return job, nil
}

func (s *service) GetBulkJobs(ctx context.Context, tenantID pgtype.UUID, jobType string) ([]db.BulkJob, error) {
	queries := db.New(s.pool)
	return queries.GetBulkJobs(ctx, db.GetBulkJobsParams{
		TenantID: tenantID,
		JobType:  jobType,
	})
}
