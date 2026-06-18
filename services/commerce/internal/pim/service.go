package pim

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"

	"commerce_modules/internal/models"
)

type InventoryContract interface {
	InitializeInventory(ctx context.Context, tx pgx.Tx, variantID uuid.UUID) error
	CascadeVariantDeletion(ctx context.Context, tx pgx.Tx, variantID uuid.UUID) error
}

type DBPool interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	db        DBPool
	repo      *Repository
	inventory InventoryContract
}

func NewService(db DBPool, repo *Repository, inventory InventoryContract) *Service {
	return &Service{
		db:        db,
		repo:      repo,
		inventory: inventory,
	}
}

func (s *Service) CreateProduct(ctx context.Context, product *models.Product) error {
	return s.repo.CreateProduct(ctx, s.db, product)
}

func (s *Service) GetProduct(ctx context.Context, orgID, tenantID, productID uuid.UUID) (*models.Product, error) {
	return s.repo.GetProduct(ctx, s.db, orgID, tenantID, productID)
}

func (s *Service) UpdateProduct(ctx context.Context, product *models.Product) error {
	return s.repo.UpdateProduct(ctx, s.db, product)
}

func (s *Service) DeleteProduct(ctx context.Context, orgID, tenantID, productID uuid.UUID) error {
	// Start transaction because deleting a product deletes variants, which need inventory updates
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Lock product to prevent concurrent modifications
	if err := s.repo.LockProduct(ctx, tx, orgID, tenantID, productID, "FOR UPDATE"); err != nil {
		return err
	}

	// Fetch all variants first to cascade delete in inventory
	variants, err := s.repo.ListVariants(ctx, tx, orgID, tenantID, productID)
	if err != nil {
		return err
	}

	for _, v := range variants {
		if s.inventory != nil {
			if err := s.inventory.CascadeVariantDeletion(ctx, tx, v.ID); err != nil {
				return err
			}
		}
	}

	if err := s.repo.DeleteProduct(ctx, tx, orgID, tenantID, productID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) CreateVariant(ctx context.Context, variant *models.ProductVariant) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.repo.LockProduct(ctx, tx, variant.OrgID, variant.TenantID, variant.ProductID, "FOR SHARE"); err != nil {
		return err
	}

	if err := s.repo.CreateVariant(ctx, tx, variant); err != nil {
		return err
	}

	if s.inventory != nil {
		if err := s.inventory.InitializeInventory(ctx, tx, variant.ID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Service) GetVariant(ctx context.Context, orgID, tenantID, variantID uuid.UUID) (*models.ProductVariant, error) {
	return s.repo.GetVariant(ctx, s.db, orgID, tenantID, variantID)
}

func (s *Service) ListVariants(ctx context.Context, orgID, tenantID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	return s.repo.ListVariants(ctx, s.db, orgID, tenantID, productID)
}

func (s *Service) UpdateVariant(ctx context.Context, variant *models.ProductVariant) error {
	return s.repo.UpdateVariant(ctx, s.db, variant)
}

func (s *Service) DeleteVariant(ctx context.Context, orgID, tenantID, variantID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if s.inventory != nil {
		if err := s.inventory.CascadeVariantDeletion(ctx, tx, variantID); err != nil {
			return err
		}
	}

	if err := s.repo.DeleteVariant(ctx, tx, orgID, tenantID, variantID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) SearchProducts(ctx context.Context, orgID, tenantID uuid.UUID, embedding pgvector.Vector, limit int) ([]*models.Product, error) {
	return s.repo.SearchProducts(ctx, s.db, orgID, tenantID, embedding, limit)
}
