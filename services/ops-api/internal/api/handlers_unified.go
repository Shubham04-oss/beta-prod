package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/ops-api/internal/unified"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

type UnifiedHandler struct {
	pool          *pgxpool.Pool
	dbQueries     *db.Queries
	svc           *unified.Service
	webhookSecret string
}

func NewUnifiedHandler(pool *pgxpool.Pool, svc *unified.Service, webhookSecret string) *UnifiedHandler {
	return &UnifiedHandler{
		pool:          pool,
		dbQueries:     db.New(pool),
		svc:           svc,
		webhookSecret: webhookSecret,
	}
}

func (h *UnifiedHandler) RegisterRoutes(protected chi.Router, public chi.Router) {
	// Protected Routes (Require Firebase JWT & Multi-Tenant Middleware)
	protected.Post("/unified/sync/push/product", h.HandlePushProduct)

	// Public Webhook Route (Uses HMAC Signature Verification instead of JWT)
	public.Post("/unified/webhook", h.HandleWebhook)
}

func (h *UnifiedHandler) HandlePushProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := authcontext.GetTenantID(r.Context())
	if err != nil {
		http.Error(w, "tenant id required", http.StatusUnauthorized)
		return
	}

	orgID, err := authcontext.GetOrgID(r.Context())
	if err != nil {
		http.Error(w, "org id required", http.StatusUnauthorized)
		return
	}

	var req struct {
		ProductID uuid.UUID `json:"productId"`
		Action    string    `json:"action"` // e.g., "UPSERT" or "DELETE"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Enqueue the push job asynchronously
	h.svc.EnqueuePushJob(tenantID, orgID, req.ProductID.String(), req.Action)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

func (h *UnifiedHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Read entire body for HMAC verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	// 2. Verify HMAC Signature
	signature := r.Header.Get("X-Unified-Signature")
	if h.webhookSecret == "" {
		http.Error(w, "webhook secret not configured", http.StatusInternalServerError)
		return
	}
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 3. Extract Connection ID
	var payload struct {
		ConnectionID string          `json:"connection_id"`
		Event        string          `json:"event"`
		Data         json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	// 4. Look up internal connection context to ensure secure tenant isolation.
	type connectionContext struct {
		id       uuid.UUID
		orgID    uuid.UUID
		tenantID uuid.UUID
		provider string
	}
	var conn connectionContext
	err = h.pool.QueryRow(r.Context(), `
		SELECT id, org_id, tenant_id, provider
		FROM commerce_connections
		WHERE unified_connection_id = $1 AND status = 'ACTIVE' AND deleted_at IS NULL
		LIMIT 1
	`, payload.ConnectionID).Scan(&conn.id, &conn.orgID, &conn.tenantID, &conn.provider)
	if err != nil {
		// Log this instead of exposing error, could be a disabled connection
		http.Error(w, "unauthorized webhook connection", http.StatusForbidden)
		return
	}

	// 5. Create a system context with all identity pillars for downstream event consumers.
	ctx := authcontext.WithTenantID(context.Background(), conn.tenantID.String())
	ctx = authcontext.WithOrgID(ctx, conn.orgID.String())
	ctx = authcontext.WithUserID(ctx, uuid.Nil.String())
	ctx = authcontext.WithRole(ctx, "SYSTEM")

	eventPayload, err := json.Marshal(map[string]interface{}{
		"connection_id":         conn.id.String(),
		"unified_connection_id": payload.ConnectionID,
		"provider":              conn.provider,
		"event":                 payload.Event,
		"data":                  json.RawMessage(payload.Data),
	})
	if err != nil {
		http.Error(w, "invalid webhook payload", http.StatusBadRequest)
		return
	}

	eventMetadata, _ := json.Marshal(map[string]string{
		"source":     "unified.to",
		"actor_id":   "00000000-0000-0000-0000-000000000000",
		"actor_role": "SYSTEM",
	})

	_, err = h.pool.Exec(ctx, `
		INSERT INTO oms_outbox_events (topic, aggregate_id, aggregate_type, tenant_id, org_id, payload, metadata)
		VALUES ($1, $2, 'commerce_connection', $3, $4, $5, $6)
	`, "unified.webhook.received", conn.id, conn.tenantID, conn.orgID, eventPayload, eventMetadata)
	if err != nil {
		http.Error(w, "failed to persist webhook event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
