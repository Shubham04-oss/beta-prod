package unified

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/authcontext"
	"go.temporal.io/sdk/client"
)

type WebhookHandler struct {
	temporalClient client.Client
	dbpool         *pgxpool.Pool
	webhookSecret  string
}

func NewWebhookHandler(tc client.Client, dbpool *pgxpool.Pool, secret string) *WebhookHandler {
	return &WebhookHandler{
		temporalClient: tc,
		dbpool:         dbpool,
		webhookSecret:  secret,
	}
}

func (h *WebhookHandler) RegisterRoutes(r chi.Router) {
	r.Post("/v1/unified/webhooks", h.HandleWebhook)
}

type UnifiedWebhookPayload struct {
	Event        string                 `json:"event"`
	ConnectionID string                 `json:"connection_id"`
	WorkspaceID  string                 `json:"workspace_id"`
	Data         map[string]interface{} `json:"data"`
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	// Verify HMAC Signature if secret is provided
	if h.webhookSecret == "" {
		http.Error(w, "webhook secret not configured", http.StatusInternalServerError)
		return
	}
	signature := r.Header.Get("X-Unified-Signature")
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload UnifiedWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.ConnectionID == "" {
		http.Error(w, "connection_id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var orgID, tenantID string
	err = h.dbpool.QueryRow(ctx, `
		SELECT org_id::text, tenant_id::text
		FROM commerce_connections
		WHERE unified_connection_id = $1
			AND status = 'ACTIVE'
			AND deleted_at IS NULL
		LIMIT 1
	`, payload.ConnectionID).Scan(&orgID, &tenantID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "connection not found", http.StatusForbidden)
			return
		}
		http.Error(w, "failed to verify connection", http.StatusInternalServerError)
		return
	}

	ctx = authcontext.WithTenantID(ctx, tenantID)
	ctx = authcontext.WithOrgID(ctx, orgID)
	ctx = authcontext.WithUserID(ctx, uuid.Nil.String())
	ctx = authcontext.WithRole(ctx, "SYSTEM")

	if payload.Event == "accounting.order.created" || payload.Event == "accounting.order.updated" {
		// Start temporal workflow to process this incoming order webhook reliably
		options := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("sync-inbound-order-%s-%s-%s", payload.ConnectionID, payload.Event, extractID(payload.Data)),
			TaskQueue: "oms-task-queue-v3",
		}

		// We dispatch to a dedicated Temporal workflow for parsing and ingesting the order
		// This workflow will idempotently map the AccountingOrder to an OMS Order
		_, err := h.temporalClient.ExecuteWorkflow(ctx, options, "SyncInboundOrderWorkflow", payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func extractID(data map[string]interface{}) string {
	if id, ok := data["id"].(string); ok {
		return id
	}
	return "unknown"
}
