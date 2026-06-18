package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

type SettingsHandlers struct {
	queries *db.Queries
}

func NewSettingsHandlers(queries *db.Queries) *SettingsHandlers {
	return &SettingsHandlers{queries: queries}
}

func (h *SettingsHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/settings/tenant", h.GetTenantSettings)
	r.Put("/api/v1/settings/tenant", h.UpsertTenantSettings)
}

type tenantSettingsRequest struct {
	InventoryAllocationModel string `json:"inventory_allocation_model"`
	AutoPoEnabled            bool   `json:"auto_po_enabled"`
	DefaultLowStockThreshold int32  `json:"default_low_stock_threshold"`
	CostingMethod            string `json:"costing_method"`
}

type tenantSettingsResponse struct {
	ID                       string `json:"id"`
	OrgID                    string `json:"org_id"`
	TenantID                 string `json:"tenant_id"`
	InventoryAllocationModel string `json:"inventory_allocation_model"`
	AutoPoEnabled            bool   `json:"auto_po_enabled"`
	DefaultLowStockThreshold int32  `json:"default_low_stock_threshold"`
	CostingMethod            string `json:"costing_method"`
	UpdatedBy                string `json:"updated_by,omitempty"`
	CreatedAt                string `json:"created_at"`
	UpdatedAt                string `json:"updated_at"`
}

func (h *SettingsHandlers) GetTenantSettings(w http.ResponseWriter, r *http.Request) {
	tenantUUID, ok := requiredTenantUUID(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}

	settings, err := h.queries.GetTenantSettings(r.Context(), db.GetTenantSettingsParams{
		TenantID: tenantUUID,
		OrgID:    orgUUID,
	})
	if err != nil {
		http.Error(w, "tenant settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(makeTenantSettingsResponse(settings))
}

func (h *SettingsHandlers) UpsertTenantSettings(w http.ResponseWriter, r *http.Request) {
	role, err := authcontext.GetRole(r.Context())
	if err != nil || strings.ToUpper(role) != "ADMIN" {
		http.Error(w, "Forbidden: Only Administrators can update tenant settings", http.StatusForbidden)
		return
	}

	tenantUUID, ok := requiredTenantUUID(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}

	var req tenantSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.InventoryAllocationModel = strings.ToUpper(strings.TrimSpace(req.InventoryAllocationModel))
	req.CostingMethod = strings.ToUpper(strings.TrimSpace(req.CostingMethod))
	if req.InventoryAllocationModel == "" {
		req.InventoryAllocationModel = "HARD"
	}
	if req.InventoryAllocationModel != "HARD" && req.InventoryAllocationModel != "SOFT" {
		http.Error(w, "inventory_allocation_model must be HARD or SOFT", http.StatusBadRequest)
		return
	}
	if req.CostingMethod == "" {
		req.CostingMethod = "WAC"
	}
	if req.CostingMethod != "WAC" && req.CostingMethod != "FIFO" {
		http.Error(w, "costing_method must be WAC or FIFO", http.StatusBadRequest)
		return
	}
	if req.DefaultLowStockThreshold < 0 {
		http.Error(w, "default_low_stock_threshold cannot be negative", http.StatusBadRequest)
		return
	}

	updatedBy := pgtype.UUID{}
	if userID, err := authcontext.GetUserID(r.Context()); err == nil {
		if parsed, err := uuid.Parse(userID); err == nil {
			updatedBy = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	settings, err := h.queries.UpsertTenantSettings(r.Context(), db.UpsertTenantSettingsParams{
		OrgID:                    orgUUID,
		TenantID:                 tenantUUID,
		InventoryAllocationModel: pgtype.Text{String: req.InventoryAllocationModel, Valid: true},
		AutoPoEnabled:            pgtype.Bool{Bool: req.AutoPoEnabled, Valid: true},
		DefaultLowStockThreshold: pgtype.Int4{Int32: req.DefaultLowStockThreshold, Valid: true},
		CostingMethod:            pgtype.Text{String: req.CostingMethod, Valid: true},
		UpdatedBy:                updatedBy,
	})
	if err != nil {
		http.Error(w, "failed to update tenant settings", http.StatusInternalServerError)
		return
	}

	details, _ := json.Marshal(map[string]interface{}{
		"inventory_allocation_model":  req.InventoryAllocationModel,
		"auto_po_enabled":             req.AutoPoEnabled,
		"default_low_stock_threshold": req.DefaultLowStockThreshold,
		"costing_method":              req.CostingMethod,
	})
	_, _ = h.queries.InsertAuditEvent(r.Context(), db.InsertAuditEventParams{
		OrgID:      orgUUID,
		TenantID:   tenantUUID,
		ActorEmail: pgtype.Text{String: auditActor(r), Valid: true},
		Action:     "TENANT_SETTINGS_UPDATED",
		EntityType: "TENANT_SETTINGS",
		EntityID:   settings.ID,
		Details:    details,
		IpAddress:  pgtype.Text{String: r.RemoteAddr, Valid: true},
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(makeTenantSettingsResponse(settings))
}

func makeTenantSettingsResponse(settings db.TenantSetting) tenantSettingsResponse {
	return tenantSettingsResponse{
		ID:                       pgUUIDString(settings.ID),
		OrgID:                    pgUUIDString(settings.OrgID),
		TenantID:                 pgUUIDString(settings.TenantID),
		InventoryAllocationModel: settings.InventoryAllocationModel.String,
		AutoPoEnabled:            settings.AutoPoEnabled.Bool,
		DefaultLowStockThreshold: settings.DefaultLowStockThreshold.Int32,
		CostingMethod:            settings.CostingMethod.String,
		UpdatedBy:                pgUUIDString(settings.UpdatedBy),
		CreatedAt:                pgTimeString(settings.CreatedAt),
		UpdatedAt:                pgTimeString(settings.UpdatedAt),
	}
}

func requiredTenantUUID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	tenantID, err := authcontext.GetTenantID(r.Context())
	if err != nil {
		http.Error(w, "tenant id required", http.StatusUnauthorized)
		return pgtype.UUID{}, false
	}
	parsed, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return pgtype.UUID{}, false
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, true
}

func requiredOrgUUID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	orgID, err := authcontext.GetOrgID(r.Context())
	if err != nil {
		http.Error(w, "org id required", http.StatusUnauthorized)
		return pgtype.UUID{}, false
	}
	parsed, err := uuid.Parse(orgID)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return pgtype.UUID{}, false
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, true
}

func pgUUIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}

func pgTimeString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func auditActor(r *http.Request) string {
	if userID, err := authcontext.GetUserID(r.Context()); err == nil && userID != "" {
		return userID
	}
	return "system"
}
