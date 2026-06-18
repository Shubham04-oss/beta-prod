package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/authcontext"
	db_sqlc "github.com/synq/pkg/db"
	unifiedgosdk "github.com/unified-to/unified-go-sdk"
	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
)

type IntegrationsHandler struct {
	db          *pgxpool.Pool
	queries     *db_sqlc.Queries
	unified     *unifiedgosdk.UnifiedTo
	workspaceID string
}

func NewIntegrationsHandler(db *pgxpool.Pool, workspaceID string) *IntegrationsHandler {
	token := os.Getenv("UNIFIED_TO_TOKEN")
	u := unifiedgosdk.New(
		unifiedgosdk.WithSecurity(token),
	)

	return &IntegrationsHandler{
		db:          db,
		queries:     db_sqlc.New(db),
		unified:     u,
		workspaceID: workspaceID,
	}
}

func (h *IntegrationsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/integrations/connections", h.HandleListConnections)
	r.Post("/api/v1/integrations/auth-url", h.HandleCreateAuthURL)
	r.Post("/api/v1/integrations/callback", h.HandleSaveConnection)
}

// HandleCreateAuthURL generates an Authorization Link for the tenant to connect their Commerce tool
func (h *IntegrationsHandler) HandleCreateAuthURL(w http.ResponseWriter, r *http.Request) {
	tenantID, err := authcontext.GetTenantID(r.Context())
	if err != nil || tenantID == "" {
		http.Error(w, "Unauthorized: missing tenant context", http.StatusUnauthorized)
		return
	}

	authBase := os.Getenv("UNIFIED_AUTH_BASE_URL")
	if authBase == "" {
		authBase = "https://api.unified.to/integration/auth"
	}
	authURL, err := url.Parse(authBase)
	if err != nil {
		http.Error(w, "Unified auth URL is misconfigured", http.StatusInternalServerError)
		return
	}
	query := authURL.Query()
	query.Set("state", tenantID)
	query.Set("workspace_id", h.workspaceID)
	query.Set("category", "commerce")
	authURL.RawQuery = query.Encode()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"auth_url": %q}`, authURL.String())))
}

type ConnectionRequest struct {
	ConnectionID string `json:"connection_id"`
	Category     string `json:"category"` // e.g., "commerce"
	Provider     string `json:"provider"`
}

// HandleSaveConnection saves the resulting Unified.to connection_id into Postgres
func (h *IntegrationsHandler) HandleSaveConnection(w http.ResponseWriter, r *http.Request) {
	tenantID, err := authcontext.GetTenantID(r.Context())
	orgID, err2 := authcontext.GetOrgID(r.Context())
	if err != nil || err2 != nil {
		http.Error(w, "Unauthorized: missing tenant context", http.StatusUnauthorized)
		return
	}

	var req ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.ConnectionID = strings.TrimSpace(req.ConnectionID)
	req.Category = strings.ToLower(strings.TrimSpace(req.Category))
	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	if req.ConnectionID == "" {
		http.Error(w, "connection_id is required", http.StatusBadRequest)
		return
	}
	if req.Category == "" {
		req.Category = "commerce"
	}

	var pgTenantID, pgOrgID pgtype.UUID
	pgTenantID.Scan(tenantID)
	pgOrgID.Scan(orgID)

	// For Commerce categories, also save to commerce_connections for Temporal Background syncing
	if req.Category == "commerce" || req.Category == "accounting" {
		provider, err := h.resolveUnifiedProvider(r.Context(), req.ConnectionID, req.Provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		_, err = h.db.Exec(r.Context(), `
			INSERT INTO commerce_connections (org_id, tenant_id, unified_connection_id, provider, status)
			VALUES ($1, $2, $3, $4, 'ACTIVE')
			ON CONFLICT (tenant_id, unified_connection_id) DO UPDATE
			SET provider = EXCLUDED.provider,
				status = 'ACTIVE',
				deleted_at = NULL,
				updated_at = CURRENT_TIMESTAMP
		`, pgOrgID, pgTenantID, req.ConnectionID, provider)
		if err != nil {
			http.Error(w, "Failed to save commerce connection", http.StatusInternalServerError)
			return
		}

		details, _ := json.Marshal(map[string]string{
			"connection_id": req.ConnectionID,
			"provider":      provider,
			"category":      req.Category,
		})
		_, _ = h.queries.InsertAuditEvent(r.Context(), db_sqlc.InsertAuditEventParams{
			OrgID:      pgOrgID,
			TenantID:   pgTenantID,
			ActorEmail: pgtype.Text{String: auditActor(r), Valid: true},
			Action:     "INTEGRATION_CONNECTED",
			EntityType: "COMMERCE_CONNECTION",
			EntityID:   pgtype.UUID{},
			Details:    details,
			IpAddress:  pgtype.Text{String: r.RemoteAddr, Valid: true},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success","message":"Integration secured"}`))
}

func (h *IntegrationsHandler) HandleListConnections(w http.ResponseWriter, r *http.Request) {
	tenantID, err := authcontext.GetTenantID(r.Context())
	orgID, orgErr := authcontext.GetOrgID(r.Context())
	if err != nil || orgErr != nil || tenantID == "" || orgID == "" {
		http.Error(w, "Unauthorized: missing tenant context", http.StatusUnauthorized)
		return
	}
	var pgTenantID pgtype.UUID
	if err := pgTenantID.Scan(tenantID); err != nil {
		http.Error(w, "Invalid tenant context", http.StatusBadRequest)
		return
	}
	var pgOrgID pgtype.UUID
	if err := pgOrgID.Scan(orgID); err != nil {
		http.Error(w, "Invalid org context", http.StatusBadRequest)
		return
	}

	connections, err := h.queries.GetCommerceConnections(r.Context(), db_sqlc.GetCommerceConnectionsParams{
		TenantID: pgTenantID,
		OrgID:    pgOrgID,
	})
	if err != nil {
		http.Error(w, "Failed to list connections", http.StatusInternalServerError)
		return
	}
	if connections == nil {
		connections = []db_sqlc.CommerceConnection{}
	}

	out := make([]connectionResponse, 0, len(connections))
	for _, conn := range connections {
		out = append(out, connectionResponse{
			ID:                  uuid.UUID(conn.ID.Bytes).String(),
			UnifiedConnectionID: conn.UnifiedConnectionID,
			Provider:            conn.Provider,
			Status:              conn.Status.String,
			CreatedAt:           timestampString(conn.CreatedAt),
			UpdatedAt:           timestampString(conn.UpdatedAt),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"connections": out})
}

type connectionResponse struct {
	ID                  string `json:"id"`
	UnifiedConnectionID string `json:"unified_connection_id"`
	Provider            string `json:"provider"`
	Status              string `json:"status"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

func (h *IntegrationsHandler) resolveUnifiedProvider(ctx context.Context, connectionID, fallback string) (string, error) {
	token := os.Getenv("UNIFIED_TO_TOKEN")
	if h.unified != nil && token != "" {
		res, err := h.unified.Unified.GetUnifiedConnection(ctx, operations.GetUnifiedConnectionRequest{ID: connectionID})
		if err == nil && res != nil && res.Connection != nil && res.Connection.IntegrationType != "" {
			return strings.ToLower(res.Connection.IntegrationType), nil
		}
	}
	if fallback != "" && os.Getenv("ENV") != "production" {
		return fallback, nil
	}
	return "", fmt.Errorf("unable to verify Unified connection provider")
}

func timestampString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}
