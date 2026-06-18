package api

import (
	"encoding/json"
	"net/http"

	"github.com/synq/pkg/authcontext"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synq/pkg/db"
)

type AuditHandlers struct {
	queries *db.Queries
}

func NewAuditHandlers(queries *db.Queries) *AuditHandlers {
	return &AuditHandlers{queries: queries}
}

func (h *AuditHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/audit", h.ListAuditEventsHandler)
}

func (h *AuditHandlers) ListAuditEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract Tenant ID from context
	tenantIDStr, err := authcontext.GetTenantID(r.Context())
	if err != nil || tenantIDStr == "" {
		http.Error(w, "Unauthorized: Missing Tenant ID", http.StatusForbidden)
		return
	}

	var tenantUUID pgtype.UUID
	if err := tenantUUID.Scan(tenantIDStr); err != nil {
		http.Error(w, "Invalid Tenant ID", http.StatusBadRequest)
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}

	events, err := h.queries.GetAuditEventsByTenant(r.Context(), db.GetAuditEventsByTenantParams{
		TenantID: tenantUUID,
		OrgID:    orgUUID,
	})
	if err != nil {
		http.Error(w, "Failed to fetch audit events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle null slice
	if events == nil {
		events = []db.GetAuditEventsByTenantRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
