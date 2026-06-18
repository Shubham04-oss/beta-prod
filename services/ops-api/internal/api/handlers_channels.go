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

type ChannelHandlers struct {
	queries *db.Queries
}

func NewChannelHandlers(queries *db.Queries) *ChannelHandlers {
	return &ChannelHandlers{queries: queries}
}

func (h *ChannelHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/channels", h.ListSalesChannels)
	r.Post("/api/v1/channels", h.CreateSalesChannel)
}

type createSalesChannelRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
	Active   *bool  `json:"active"`
}

type salesChannelResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *ChannelHandlers) ListSalesChannels(w http.ResponseWriter, r *http.Request) {
	tenantUUID, ok := requiredTenantUUID(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}

	channels, err := h.queries.ListSalesChannels(r.Context(), db.ListSalesChannelsParams{
		TenantID: tenantUUID,
		OrgID:    orgUUID,
	})
	if err != nil {
		http.Error(w, "failed to list sales channels", http.StatusInternalServerError)
		return
	}
	if channels == nil {
		channels = []db.SalesChannel{}
	}

	out := make([]salesChannelResponse, 0, len(channels))
	for _, channel := range channels {
		out = append(out, makeSalesChannelResponse(channel))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"channels": out})
}

func (h *ChannelHandlers) CreateSalesChannel(w http.ResponseWriter, r *http.Request) {
	role, err := authcontext.GetRole(r.Context())
	if err != nil || (strings.ToUpper(role) != "ADMIN" && strings.ToUpper(role) != "EDITOR") {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
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

	var req createSalesChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if len(req.Currency) != 3 {
		http.Error(w, "currency must be a 3-letter ISO code", http.StatusBadRequest)
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	channel, err := h.queries.CreateSalesChannel(r.Context(), db.CreateSalesChannelParams{
		OrgID:    orgUUID,
		TenantID: tenantUUID,
		Name:     req.Name,
		Currency: pgtype.Text{String: req.Currency, Valid: true},
		IsActive: pgtype.Bool{Bool: active, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create sales channel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(makeSalesChannelResponse(channel))
}

func makeSalesChannelResponse(channel db.SalesChannel) salesChannelResponse {
	return salesChannelResponse{
		ID:        uuid.UUID(channel.ID.Bytes).String(),
		Name:      channel.Name,
		Currency:  channel.Currency.String,
		Active:    channel.IsActive.Bool,
		CreatedAt: channelTimeString(channel.CreatedAt),
		UpdatedAt: channelTimeString(channel.UpdatedAt),
	}
}

func channelTimeString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}
