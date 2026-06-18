package unified

import (
	"commerce_modules/internal/models"
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrSyncFailed      = errors.New("sync failed")
)

type PIMClient interface {
	GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error)
	// Some operations might need the full product or all variants
}

// In PIM adapter, GetVariant also acts as GetProduct returning the first variant if productID is passed. Let's define the local PIMClient interface.
type CustomPIMClient interface {
	GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error)
	GetProduct(ctx context.Context, tenantID, orgID, productID uuid.UUID) (*models.Product, error)
	ListVariants(ctx context.Context, tenantID, orgID, productID uuid.UUID) ([]*models.ProductVariant, error)
}

type OMSClient interface {
	CreateOrder(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error
	GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error)
}

type InventoryClient interface {
	AdjustStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string, quantity int) error
	GetStock(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID, productID string) (int, error)
}

type SyncService struct {
	unifiedClient UnifiedClient
	pimClient     CustomPIMClient
	omsClient     OMSClient
	invClient     InventoryClient

	syncJobs map[string]string
	jobsMu   sync.RWMutex
	keyLocks keyMutex
}

type keyMutex struct {
	locks [256]sync.Mutex
}

func (k *keyMutex) Lock(key string) func() {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() % 256
	k.locks[idx].Lock()
	return k.locks[idx].Unlock
}

func NewSyncService(
	uc UnifiedClient,
	pim CustomPIMClient,
	oms OMSClient,
	inv InventoryClient,
) *SyncService {
	return &SyncService{
		unifiedClient: uc,
		pimClient:     pim,
		omsClient:     oms,
		invClient:     inv,
		syncJobs:      make(map[string]string),
	}
}

func (s *SyncService) GenerateJobID() string {
	return "job-" + uuid.New().String()
}

func (s *SyncService) SetSyncStatus(jobID, status string) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	s.syncJobs[jobID] = status
}

func (s *SyncService) GetSyncStatus(jobID string) string {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	if status, ok := s.syncJobs[jobID]; ok {
		return status
	}
	return "not_found"
}

func (s *SyncService) PushProduct(ctx context.Context, tenantID, orgID uuid.UUID, connectionID string, productID string) error {
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return err
	}

	// Fetch the actual product
	product, err := s.pimClient.GetProduct(ctx, tenantID, orgID, prodUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrProductNotFound
		}
		return err
	}

	// Fetch all variants for the product
	variants, err := s.pimClient.ListVariants(ctx, tenantID, orgID, prodUUID)
	if err != nil {
		return fmt.Errorf("failed to list variants: %w", err)
	}

	payload := MapProductToUnified(product, variants)

	_, err = s.unifiedClient.PushProduct(ctx, connectionID, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSyncFailed, err)
	}

	return nil
}

func (s *SyncService) PullOrder(ctx context.Context, tenantID, orgID uuid.UUID, connectionID string, orderID string) (*models.Order, error) {
	payload, err := s.unifiedClient.PullOrder(ctx, connectionID, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSyncFailed, err)
	}

	order, items, err := MapUnifiedToOrder(payload, tenantID, orgID)
	if err != nil {
		return nil, err
	}

	// Lock during the idempotency check and creation to avoid TOCTOU
	unlock := s.keyLocks.Lock(order.ID.String())
	defer unlock()

	// Idempotency: check if exists
	existingOrder, err := s.omsClient.GetOrder(ctx, tenantID, orgID, order.ID)
	if existingOrder != nil {
		return existingOrder, nil
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing order: %w", err)
	}

	err = s.omsClient.CreateOrder(ctx, tenantID, orgID, order, items)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *SyncService) HandleWebhook(ctx context.Context, event string, payload map[string]interface{}) error {
	if event != "inventory.updated" {
		return nil // Ignore other events
	}

	productID, quantity, err := ExtractInventoryUpdate(payload)
	if err != nil {
		return err
	}

	unlock := s.keyLocks.Lock(productID)
	defer unlock()

	tenantID, _ := ctx.Value("tenantID").(uuid.UUID)
	orgID, _ := ctx.Value("orgID").(uuid.UUID)

	currentStock, err := s.invClient.GetStock(ctx, tenantID, orgID, productID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			currentStock = 0
		} else {
			return fmt.Errorf("failed to get stock: %w", err)
		}
	}

	diff := quantity - currentStock
	return s.invClient.AdjustStock(ctx, tenantID, orgID, productID, diff)
}
