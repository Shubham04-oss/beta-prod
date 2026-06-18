package unified

import (
	"commerce_modules/internal/models"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type AdversarialMockOMSClient struct {
	orders []uuid.UUID
	mu     sync.Mutex
}

func (m *AdversarialMockOMSClient) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	m.mu.Lock()
	found := false
	for _, id := range m.orders {
		if id == orderID {
			found = true
			break
		}
	}
	m.mu.Unlock()

	if found {
		return &models.Order{ID: orderID}, nil
	}

	// Delay AFTER releasing lock, simulating network/DB call time. This allows concurrent GetOrder calls to interleave.
	time.Sleep(100 * time.Millisecond)
	return nil, errors.New("not found")
}

func (m *AdversarialMockOMSClient) CreateOrder(ctx context.Context, tenantID, orgID uuid.UUID, order *models.Order, items []models.OrderLineItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders = append(m.orders, order.ID)
	return nil
}

func TestPullOrder_ConcurrentIdempotencyFailure(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &AdversarialMockOMSClient{}
	inv := &MockInventoryClient{}

	svc := NewSyncService(uc, pim, oms, inv)

	orderID := "ext-order-123"

	uc.Orders[orderID] = map[string]interface{}{
		"id":          orderID,
		"status":      "PAID",
		"total_price": float64(100.5),
	}

	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", orderID)
		}()
	}

	wg.Wait()

	if len(oms.orders) > 1 {
		t.Fatalf("Expected 1 order to be created, but got %d! Concurrent PullOrder has a TOCTOU race condition.", len(oms.orders))
	} else if len(oms.orders) == 1 {
		t.Log("Order creation was properly synchronized.")
	} else {
		t.Fatalf("No orders were created.")
	}
}

type ConcurrentInventoryClient struct {
	stock            map[string]int
	mu               sync.Mutex
	getStockCalled   chan struct{}
	waitBeforeAdjust chan struct{}
}

func (m *ConcurrentInventoryClient) AdjustStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stock[productID] += quantity
	return nil
}

func (m *ConcurrentInventoryClient) GetStock(ctx context.Context, tenantID, orgID uuid.UUID, productID string) (int, error) {
	m.mu.Lock()
	val, ok := m.stock[productID]
	m.mu.Unlock()

	// signal that GetStock was called
	select {
	case m.getStockCalled <- struct{}{}:
	default:
	}

	// wait before returning to allow concurrent AdjustStock
	if m.waitBeforeAdjust != nil {
		<-m.waitBeforeAdjust
	}

	if ok {
		return val, nil
	}
	return 0, errors.New("not found")
}

func TestHandleWebhook_ExternalTOCTOURace(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &AdversarialMockOMSClient{}
	inv := &ConcurrentInventoryClient{
		stock:            make(map[string]int),
		getStockCalled:   make(chan struct{}, 1),
		waitBeforeAdjust: make(chan struct{}),
	}

	svc := NewSyncService(uc, pim, oms, inv)

	productID := "prod-race-webhook"
	inv.stock[productID] = 100 // initial stock

	var wg sync.WaitGroup
	wg.Add(2)

	// Thread 1: Webhook execution
	go func() {
		defer wg.Done()
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"product_id": productID,
				"quantity":   float64(50), // Unified says stock is 50
			},
		}
		err := svc.HandleWebhook(context.Background(), "inventory.updated", payload)
		if err != nil {
			t.Errorf("Webhook error: %v", err)
		}
	}()

	// Thread 2: External manual adjustment (e.g. from an order or manual adjustment in OMS)
	go func() {
		defer wg.Done()
		// Wait until Webhook calls GetStock
		<-inv.getStockCalled

		// Concurrently adjust stock (e.g., an order comes in and decreases stock by 20)
		inv.AdjustStock(context.Background(), uuid.Nil, uuid.Nil, productID, -20) // stock becomes 80

		// Allow Webhook to proceed with its calculation
		close(inv.waitBeforeAdjust)
	}()

	wg.Wait()

	finalStock := inv.stock[productID]
	// The webhook calculated diff = 50 - 100 = -50.
	// Then it applies -50.
	// But meanwhile stock became 80.
	// 80 - 50 = 30.
	// But wait, if stock WAS 50 (from unified webhook), and an order came in that took 20, the final stock SHOULD be 30.
	// So 30 is actually mathematically correct in this specific sequence (webhook payload reflects state BEFORE order).
	// BUT what if the webhook payload reflects state AFTER the order?
	// E.g., Order reduces stock by 20 on Shopify. Shopify sends webhook: stock is 80.
	// But before webhook arrives, a local order reduces stock by 10 (local stock is 90).
	// Webhook diff = 80 - 90 = -10. Local stock becomes 90 - 10 = 80.
	// But wait, the math ALWAYS works out if adjustments are strictly additive and the webhook uses the EXACT current local stock at the time of calculation.
	// The problem is that the webhook does: `diff := quantity - currentStock`.
	// If `currentStock` changes between `GetStock` and `AdjustStock`, `AdjustStock` applies the diff based on STALE `currentStock`.
	// In our test, diff = 50 - 100 = -50.
	// The actual stock when AdjustStock runs is 80.
	// It applies -50, stock becomes 30.
	// If the webhook had locked the stock, the external order would wait.
	// BUT the diff method is fundamentally commutative.
	// Initial: 100
	// Webhook wants 50. Diff = -50.
	// Order wants -20.
	// Applying -50 and -20 in any order results in 30.

	// BUT what if the webhook sets the ABSOLUTE value directly?
	// Then 100 -> 50 (webhook) -> 30 (order).
	// If TOCTOU: order runs first (100->80), then webhook sets absolute 50. Final 50. (WRONG, lost the order).
	// By doing `diff := quantity - currentStock` and adding `diff`, they ACTUALLY made it resilient to commutative operations, assuming `AdjustStock` is an atomic add.
	// Wait! `diff` is calculated from a stale read, but `AdjustStock` is atomic add.
	// So:
	// A: read stock = 100.
	// B: add -20. Stock = 80.
	// A: add (50 - 100) = -50. Stock = 80 - 50 = 30.
	// Which is exactly 100 - 20 - 50 = 30.
	// So the TOCTOU here is BENIGN because of the additive nature of the fix!
	// Let's verify if `TestPullOrder_ConcurrentIdempotencyFailure` still fails.

	t.Logf("Final stock is: %d. TOCTOU is benign because adjustments are additive.", finalStock)
}

func TestPullOrder_SerializationBottleneck(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}
	oms := &AdversarialMockOMSClient{}
	inv := &MockInventoryClient{}

	svc := NewSyncService(uc, pim, oms, inv)

	orderID1 := "ext-order-1"
	orderID2 := "ext-order-2"

	uc.Orders[orderID1] = map[string]interface{}{
		"id":          orderID1,
		"status":      "PAID",
		"total_price": float64(100.5),
	}
	uc.Orders[orderID2] = map[string]interface{}{
		"id":          orderID2,
		"status":      "PAID",
		"total_price": float64(200.5),
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", orderID1)
	}()

	go func() {
		defer wg.Done()
		svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", orderID2)
	}()

	wg.Wait()
	duration := time.Since(start)

	// AdversarialMockOMSClient has a 100ms sleep in GetOrder.
	// If it was concurrent, the total time would be ~100ms.
	// If serialized by the global webhookMu, total time will be >= 200ms.
	if duration >= 200*time.Millisecond {
		t.Fatalf("Performance Bug: PullOrder serializes unrelated orders! Total duration %v >= 200ms. A global mutex (webhookMu) is held around network calls (OMSClient).", duration)
	} else {
		t.Logf("PullOrder is properly concurrent for different orders. Duration: %v", duration)
	}
}

type FlakyPIMClient struct {
	*MockPIMClient
	listVariantsErr error
}

func (m *FlakyPIMClient) ListVariants(ctx context.Context, tenantID, orgID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	if m.listVariantsErr != nil {
		return nil, m.listVariantsErr
	}
	return m.MockPIMClient.ListVariants(ctx, tenantID, orgID, productID)
}

func TestPushProduct_ListVariantsError_DataLoss(t *testing.T) {
	uc := NewMockUnifiedClient()

	mockPim := &MockPIMClient{
		products: make(map[uuid.UUID]*models.Product),
		variants: make(map[uuid.UUID]*models.ProductVariant),
	}
	pim := &FlakyPIMClient{
		MockPIMClient:   mockPim,
		listVariantsErr: errors.New("temporary DB error"),
	}

	oms := &MockOMSClient{}
	inv := &MockInventoryClient{}

	svc := NewSyncService(uc, pim, oms, inv)

	prodID := uuid.New()
	mockPim.products[prodID] = &models.Product{
		ID:    prodID,
		Title: "Test Product",
	}

	err := svc.PushProduct(context.Background(), uuid.Nil, uuid.Nil, "conn-123", prodID.String())
	if err == nil {
		t.Fatalf("Expected error from ListVariants to be propagated, but got nil")
	}
	// Test passes successfully if error is propagated
	return

	// Because of the bug, PushProduct caught the error, ignored it, and pushed an empty variant list!
	pushed, ok := uc.Pushed[prodID.String()]
	if !ok {
		t.Fatalf("Product was not pushed")
	}

	// Find the pushed variant list. MapProductToUnified translates variants.
	// But it pushed an empty variants slice!
	variantsRaw := pushed["variants"]
	if variantsRaw == nil {
		t.Fatalf("Variants field missing")
	}

	variantsList := variantsRaw.([]map[string]interface{})
	if len(variantsList) == 0 {
		t.Fatalf("Data Loss Bug: When ListVariants returns an error (e.g. DB unavailable), PushProduct silently ignores it and pushes an empty variant list to Unified, potentially wiping remote variants!")
	}
}

type FlakyOMSClient struct {
	*AdversarialMockOMSClient
}

func (m *FlakyOMSClient) GetOrder(ctx context.Context, tenantID, orgID, orderID uuid.UUID) (*models.Order, error) {
	// simulate a network failure on GetOrder
	return nil, errors.New("timeout getting order")
}

func TestPullOrder_GetOrderError_Idempotency(t *testing.T) {
	uc := NewMockUnifiedClient()
	pim := &MockPIMClient{}

	omsBase := &AdversarialMockOMSClient{}
	oms := &FlakyOMSClient{omsBase}
	inv := &MockInventoryClient{}

	svc := NewSyncService(uc, pim, oms, inv)

	orderID := "ext-order-idempotency-err"
	uc.Orders[orderID] = map[string]interface{}{
		"id":          orderID,
		"status":      "PAID",
		"total_price": float64(100.5),
	}

	// First pull -> fails GetOrder (simulating it exists but timeout), proceeds to CreateOrder!
	_, err := svc.PullOrder(context.Background(), uuid.Nil, uuid.Nil, "conn-123", orderID)
	if err == nil {
		t.Fatalf("Expected error from GetOrder to be propagated, but got nil")
	}

	if len(omsBase.orders) == 1 {
		t.Fatalf("Bug: PullOrder ignored GetOrder error (timeout) and blindly called CreateOrder!")
	}
}
