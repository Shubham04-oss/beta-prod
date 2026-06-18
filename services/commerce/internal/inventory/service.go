package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidQuantity   = errors.New("invalid quantity")
)

type Service interface {
	AdjustStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string, quantity int) error
	ReserveStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string, quantity int) error
	ReleaseStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string, quantity int) error
	GetStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string) (int, error)
}

func getVariantID(productID string) uuid.UUID {
	uid, err := uuid.Parse(productID)
	if err == nil {
		return uid
	}
	return uuid.NewMD5(uuid.NameSpaceOID, []byte(productID))
}

type pgService struct {
	db *pgxpool.Pool
}

func NewPgService(db *pgxpool.Pool) Service {
	return &pgService{
		db: db,
	}
}

func (s *pgService) ensureLocationAndLevel(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, orgID uuid.UUID, variantID uuid.UUID) (uuid.UUID, error) {
	locID := uuid.NewMD5(uuid.NameSpaceOID, []byte(fmt.Sprintf("default-location-%s-%s", orgID.String(), tenantID.String())))

	_, err := tx.Exec(ctx, `
		INSERT INTO locations (id, org_id, tenant_id, name, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, locID, orgID, tenantID, "Default Location", "WAREHOUSE", time.Now(), time.Now())
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to ensure location: %w", err)
	}

	levelID := uuid.NewMD5(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-%s-%s-%s", orgID.String(), tenantID.String(), locID.String(), variantID.String())))
	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_levels (id, org_id, tenant_id, variant_id, location_id, available_quantity, reserved_quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`, levelID, orgID, tenantID, variantID, locID, 0, 0, time.Now(), time.Now())
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to ensure inventory level: %w", err)
	}

	return levelID, nil
}

func (s *pgService) AdjustStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string, quantity int) error {
	variantID := getVariantID(productID)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = s.ensureLocationAndLevel(ctx, tx, tenantID, orgID, variantID)
	if err != nil {
		return fmt.Errorf("failed to ensure location: %w", err)
	}

	res, err := tx.Exec(ctx, `
		UPDATE inventory_levels
		SET available_quantity = available_quantity + $1, updated_at = $2
		WHERE org_id = $3 AND tenant_id = $4 AND variant_id = $5
		AND (available_quantity + $1) >= 0
	`, quantity, time.Now(), orgID, tenantID, variantID)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInsufficientStock
	}

	return tx.Commit(ctx)
}

func (s *pgService) ReserveStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	variantID := getVariantID(productID)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = s.ensureLocationAndLevel(ctx, tx, tenantID, orgID, variantID)
	if err != nil {
		return fmt.Errorf("failed to ensure location: %w", err)
	}

	res, err := tx.Exec(ctx, `
		UPDATE inventory_levels
		SET available_quantity = available_quantity - $1,
		    reserved_quantity = reserved_quantity + $1,
		    updated_at = $2
		WHERE org_id = $3 AND tenant_id = $4 AND variant_id = $5
		AND available_quantity >= $1
	`, quantity, time.Now(), orgID, tenantID, variantID)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInsufficientStock
	}

	return tx.Commit(ctx)
}

func (s *pgService) ReleaseStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	variantID := getVariantID(productID)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = s.ensureLocationAndLevel(ctx, tx, tenantID, orgID, variantID)
	if err != nil {
		return fmt.Errorf("failed to ensure location: %w", err)
	}

	res, err := tx.Exec(ctx, `
		UPDATE inventory_levels
		SET available_quantity = available_quantity + $1,
		    reserved_quantity = reserved_quantity - $1,
		    updated_at = $2
		WHERE org_id = $3 AND tenant_id = $4 AND variant_id = $5
		AND reserved_quantity >= $1
	`, quantity, time.Now(), orgID, tenantID, variantID)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInsufficientStock
	}

	return tx.Commit(ctx)
}

func (s *pgService) GetStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string) (int, error) {
	variantID := getVariantID(productID)

	var avail int
	err := s.db.QueryRow(ctx, `
		SELECT available_quantity
		FROM inventory_levels
		WHERE org_id = $1 AND tenant_id = $2 AND variant_id = $3
	`, orgID, tenantID, variantID).Scan(&avail)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return avail, nil
}
