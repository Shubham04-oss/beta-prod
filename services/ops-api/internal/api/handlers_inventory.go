package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

type InventoryHandlers struct {
	dbQueries *db.Queries
}

func NewInventoryHandlers(dbQueries *db.Queries) *InventoryHandlers {
	return &InventoryHandlers{
		dbQueries: dbQueries,
	}
}

func (h *InventoryHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/inventory", func(r chi.Router) {
		r.Post("/adjust", h.HandleAdjustInventory)
	})
}

func (h *InventoryHandlers) HandleAdjustInventory(w http.ResponseWriter, r *http.Request) {
	tenantIDStr, err := authcontext.GetTenantID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgIDStr, _ := authcontext.GetOrgID(r.Context())

	tenantUUID, _ := uuid.Parse(tenantIDStr)
	orgUUID, _ := uuid.Parse(orgIDStr)

	var req struct {
		VariantID     string `json:"variant_id"`
		LocationID    string `json:"location_id"`
		QuantityDelta int32  `json:"quantity_delta"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	varUUID, err := uuid.Parse(req.VariantID)
	if err != nil {
		http.Error(w, "Invalid variant_id", http.StatusBadRequest)
		return
	}

	locUUID, err := uuid.Parse(req.LocationID)
	if err != nil {
		http.Error(w, "Invalid location_id", http.StatusBadRequest)
		return
	}

	ledger, err := h.dbQueries.CreateInventoryLedgerEntry(r.Context(), db.CreateInventoryLedgerEntryParams{
		OrgID:           pgtype.UUID{Bytes: orgUUID, Valid: true},
		TenantID:        pgtype.UUID{Bytes: tenantUUID, Valid: true},
		VariantID:       pgtype.UUID{Bytes: varUUID, Valid: true},
		LocationID:      pgtype.UUID{Bytes: locUUID, Valid: true},
		TransactionType: "adjustment", // A generic adjustment
		QuantityDelta:   req.QuantityDelta,
		UnitCost:        pgtype.Numeric{Int: nil, Valid: true}, // Default cost
		Notes:           pgtype.Text{String: "API Adjustment", Valid: true},
		CreatedBy:       pgtype.UUID{Bytes: uuid.Nil, Valid: false}, // System or mock user
	})
	if err != nil {
		http.Error(w, "Failed to adjust inventory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ledger)
}
