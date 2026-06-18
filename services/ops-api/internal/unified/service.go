package unified

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/ops-api/internal/telemetry"
	"github.com/synq/pkg/db"
	sdk "github.com/unified-to/unified-go-sdk"
	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
	"golang.org/x/time/rate"
)

type Service struct {
	pool       *pgxpool.Pool
	unifiedSDK *sdk.UnifiedTo
	dbQueries  *db.Queries
	workerPool *WorkerPool
	limiter    *rate.Limiter
}

func NewService(pool *pgxpool.Pool, sdkToken string) *Service {
	// Initialize Unified.to SDK with optional mock URL override
	var s *sdk.UnifiedTo
	if mockURL := os.Getenv("UNIFIED_API_URL"); mockURL != "" {
		s = sdk.New(sdk.WithSecurity(sdkToken), sdk.WithServerURL(mockURL))
	} else {
		s = sdk.New(sdk.WithSecurity(sdkToken))
	}

	// Configurable rate limit for stress testing (default 10)
	rateLimit := 10
	if limitStr := os.Getenv("UNIFIED_RATE_LIMIT"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			rateLimit = l
		}
	}

	svc := &Service{
		pool:       pool,
		unifiedSDK: s,
		dbQueries:  db.New(pool),
		limiter:    rate.NewLimiter(rate.Every(1*time.Second), rateLimit),
	}

	// Initialize Worker Pool with 5 concurrent workers and queue size of 1000
	svc.workerPool = NewWorkerPool(svc, 5, 1000)

	return svc
}

func (s *Service) StartWorkerPool(ctx context.Context) {
	s.workerPool.Start(ctx)
}

func (s *Service) SDK() *sdk.UnifiedTo {
	return s.unifiedSDK
}

func (s *Service) StopWorkerPool() {
	s.workerPool.Stop()
}

func (s *Service) EnqueuePushJob(tenantID, orgID, productID, action string) {
	s.workerPool.Enqueue(Job{
		TenantID:  tenantID,
		OrgID:     orgID,
		ProductID: productID,
		Action:    action,
	})
}

func (s *Service) ProcessPush(ctx context.Context, tenantIDStr, orgIDStr, productIDStr, action string) error {
	tenantID, _ := uuid.Parse(tenantIDStr)
	orgID, _ := uuid.Parse(orgIDStr)
	productID, _ := uuid.Parse(productIDStr)

	// Enforce RLS by setting current tenant and org.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", tenantID.String())
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String())
	if err != nil {
		return err
	}

	txQueries := s.dbQueries.WithTx(tx)

	// 1. Fetch Product with variants
	rows, err := txQueries.GetProductWithVariants(ctx, pgtype.UUID{Bytes: productID, Valid: true})
	if err != nil || len(rows) == 0 {
		return fmt.Errorf("GetProductWithVariants failed: %w", err)
	}

	product := db.Product{
		ID:    rows[0].ProductID,
		Title: rows[0].ProductTitle,
	}

	var variants []db.ProductVariant
	for _, r := range rows {
		if r.VariantID.Valid {
			variants = append(variants, db.ProductVariant{
				ID:    r.VariantID,
				Sku:   r.Sku,
				Price: r.Price,
			})
		}
	}

	// 3. Map to Unified Commerce Item
	mappedItem := MapProductToUnified(product, variants)

	// 4. Get active connections for tenant
	connRows, err := tx.Query(ctx, "SELECT id, unified_connection_id FROM commerce_connections WHERE tenant_id = $1 AND org_id = $2 AND status = 'ACTIVE' AND deleted_at IS NULL", tenantID, orgID)
	if err != nil {
		return err
	}
	defer connRows.Close()

	type connection struct {
		ID                  uuid.UUID
		UnifiedConnectionID string
	}
	var connections []connection
	for connRows.Next() {
		var c connection
		if err := connRows.Scan(&c.ID, &c.UnifiedConnectionID); err == nil {
			connections = append(connections, c)
		}
	}

	if len(connections) == 0 {
		return nil // No active connections, nothing to push
	}

	// 5. Push via Unified SDK with Exponential Backoff
	for _, conn := range connections {
		op := func() error {
			// Enforce Rate Limiter before calling API
			err := s.limiter.Wait(ctx)
			if err != nil {
				return err
			}

			if action == "DELETE" {
				// We need the Unified API ID to delete
				var externalItemID string
				err := tx.QueryRow(ctx, "SELECT unified_item_id FROM commerce_item_mappings WHERE connection_id = $1 AND product_id = $2", conn.ID, productID).Scan(&externalItemID)
				if err != nil {
					return nil // If mapping doesn't exist, it was never pushed or already deleted
				}

				req := operations.RemoveCommerceItemRequest{
					ConnectionID: conn.UnifiedConnectionID,
					ID:           externalItemID,
				}
				_, err = s.unifiedSDK.Commerce.RemoveCommerceItem(ctx, req)
				if err == nil {
					_, _ = tx.Exec(ctx, "DELETE FROM commerce_item_mappings WHERE connection_id = $1 AND product_id = $2", conn.ID, productID)
				}
				return err
			}

			// UPSERT Action
			req := operations.CreateCommerceItemRequest{
				ConnectionID: conn.UnifiedConnectionID,
				CommerceItem: *mappedItem,
			}

			res, err := s.unifiedSDK.Commerce.CreateCommerceItem(ctx, req)
			if err != nil {
				return err // Returns error to backoff
			}
			if res.StatusCode >= 400 && res.StatusCode < 500 && res.StatusCode != 429 {
				return backoff.Permanent(fmt.Errorf("client error %d", res.StatusCode)) // Do not retry client errors except 429
			}
			if res.StatusCode >= 500 {
				return fmt.Errorf("server error %d", res.StatusCode) // Will retry
			}

			if res.CommerceItem != nil && res.CommerceItem.ID != nil {
				_, err = tx.Exec(ctx, `
					INSERT INTO commerce_item_mappings (org_id, tenant_id, connection_id, product_id, unified_item_id) 
					VALUES ($1, $2, $3, $4, $5) 
					ON CONFLICT (connection_id, product_id) DO UPDATE SET unified_item_id = EXCLUDED.unified_item_id, updated_at = NOW()`,
					orgID, tenantID, conn.ID, productID, *res.CommerceItem.ID,
				)
				if err != nil {
					return fmt.Errorf("failed to save mapping: %w", err)
				}
			}

			return nil
		}

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second // Retry for up to 30s

		err := backoff.Retry(op, backoff.WithContext(b, ctx))
		if err != nil {
			log.Printf("Push failed permanently for connection %s: %v", conn.UnifiedConnectionID, err)
			telemetry.UnifiedSyncFailedTotal.WithLabelValues(tenantIDStr, conn.ID.String(), action).Inc()
			telemetry.UnifiedSyncDLQTotal.WithLabelValues(tenantIDStr, conn.ID.String(), action).Inc()

			// Record to DLQ
			payloadBytes, _ := json.Marshal(mappedItem)
			tx.Exec(ctx, `INSERT INTO sync_failures_dlq (org_id, tenant_id, connection_id, entity_type, entity_id, payload, error_message)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				orgID, tenantID, conn.ID, "product", productID, payloadBytes, err.Error())
		} else {
			telemetry.UnifiedSyncSuccessTotal.WithLabelValues(tenantIDStr, conn.ID.String(), action).Inc()
		}
	}

	return tx.Commit(ctx)
}

// SyncOrderStatusToChannel pushes real fulfillment tracking data to the connected commerce channel.
func (s *Service) SyncOrderStatusToChannel(ctx context.Context, orderID, tenantIDStr, orgIDStr string) error {
	tenantID, tenantErr := uuid.Parse(tenantIDStr)
	orgID, orgErr := uuid.Parse(orgIDStr)
	if tenantErr != nil || orgErr != nil {
		return fmt.Errorf("invalid tenant or org uuid")
	}

	ordUUID, err1 := uuid.Parse(orderID)
	if err1 != nil {
		return fmt.Errorf("invalid orderID uuid")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", tenantID.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return err
	}
	txQueries := db.New(tx)

	var orderExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM orders
			WHERE id = $1 AND tenant_id = $2 AND org_id = $3 AND deleted_at IS NULL
		)
	`, ordUUID, tenantID, orgID).Scan(&orderExists); err != nil {
		return err
	}
	if !orderExists {
		return nil
	}

	mapping, err := txQueries.GetCommerceOrderMapping(ctx, db.GetCommerceOrderMappingParams{
		OrderID:  pgtype.UUID{Bytes: ordUUID, Valid: true},
		TenantID: pgtype.UUID{Bytes: tenantID, Valid: true},
		OrgID:    pgtype.UUID{Bytes: orgID, Valid: true},
	})
	if err != nil {
		// Not a synced order, skip
		return nil
	}

	order, err := txQueries.GetOrder(ctx, db.GetOrderParams{
		ID:       pgtype.UUID{Bytes: ordUUID, Valid: true},
		TenantID: pgtype.UUID{Bytes: tenantID, Valid: true},
		OrgID:    pgtype.UUID{Bytes: orgID, Valid: true},
	})
	if err != nil {
		return err
	}

	// Only push if status is fulfilled
	if order.Status != db.OrderStatusFulfilled {
		return nil
	}

	var connIDStr string
	var provider string
	err = tx.QueryRow(ctx, `
		SELECT unified_connection_id, provider
		FROM commerce_connections
		WHERE id = $1 AND tenant_id = $2 AND org_id = $3 AND status = 'ACTIVE' AND deleted_at IS NULL
	`, mapping.ConnectionID, tenantID, orgID).Scan(&connIDStr, &provider)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}

	var carrier pgtype.Text
	var trackingNumber pgtype.Text
	err = tx.QueryRow(ctx, `
		SELECT carrier, tracking_number
		FROM fulfillments
		WHERE order_id = $1
			AND tenant_id = $2
			AND org_id = $3
			AND status IN ('shipped', 'delivered')
			AND tracking_number IS NOT NULL
			AND tracking_number <> ''
			AND cancelled_at IS NULL
		ORDER BY shipped_at DESC NULLS LAST, updated_at DESC
		LIMIT 1
	`, ordUUID, tenantID, orgID).Scan(&carrier, &trackingNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}
	if !trackingNumber.Valid || strings.TrimSpace(trackingNumber.String) == "" {
		return nil
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	path := ""
	payload := map[string]any{}
	switch {
	case strings.Contains(strings.ToLower(provider), "shopify"):
		path = fmt.Sprintf("/orders/%s/fulfillments.json", mapping.UnifiedOrderID)
		fulfillment := map[string]any{
			"tracking_number": strings.TrimSpace(trackingNumber.String),
		}
		if carrier.Valid && strings.TrimSpace(carrier.String) != "" {
			fulfillment["tracking_company"] = strings.TrimSpace(carrier.String)
		}
		payload = map[string]any{
			"fulfillment": fulfillment,
		}
	default:
		log.Printf("Skipping fulfillment sync for unsupported provider %s on order %s", provider, orderID)
		return nil
	}

	if path == "" {
		return nil
	}
	req := operations.CreatePassthroughJSONRequest{
		ConnectionID: connIDStr,
		Path:         path,
		RequestBody:  payload,
	}

	_, err = s.unifiedSDK.Passthrough.CreatePassthroughJSON(ctx, req)
	if err != nil {
		return fmt.Errorf("passthrough fulfillment push failed: %w", err)
	}

	return nil
}
