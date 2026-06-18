package api

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/synq/pkg/authcontext"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
	"github.com/synq/ops-api/internal/service"
)

type WebSocketHandler struct {
	ticketStore *service.WSTicketStore
}

func NewWebSocketHandler(ticketStore *service.WSTicketStore) *WebSocketHandler {
	return &WebSocketHandler{ticketStore: ticketStore}
}

func (h *WebSocketHandler) RegisterRoutes(r chi.Router, protected chi.Router) {
	protected.Post("/api/v1/tickets", h.GenerateTicket)
	r.Get("/api/v1/agent-stream", h.AgentStream)
}

type GenerateTicketResponse struct {
	TicketID string `json:"ticket_id"`
}

func (h *WebSocketHandler) GenerateTicket(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := authcontext.GetTenantID(r.Context())
	orgID, _ := authcontext.GetOrgID(r.Context())
	userID, _ := authcontext.GetUserID(r.Context())
	role, _ := authcontext.GetRole(r.Context())

	claims := service.TicketClaims{
		TenantID: tenantID,
		OrgID:    orgID,
		UserID:   userID,
		Role:     role,
	}

	ticketID, err := h.ticketStore.CreateTicket(r.Context(), claims)
	if err != nil {
		log.Printf("Failed to create WS ticket: %v", err)
		http.Error(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateTicketResponse{TicketID: ticketID})
}

type AGUIIntent struct {
	Type      string      `json:"type"`
	Component string      `json:"component,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func (h *WebSocketHandler) AgentStream(w http.ResponseWriter, r *http.Request) {
	ticketID := r.URL.Query().Get("ticket")
	if ticketID == "" {
		http.Error(w, "Missing ticket", http.StatusUnauthorized)
		return
	}

	claims, err := h.ticketStore.ConsumeTicket(r.Context(), ticketID)
	if err != nil {
		log.Printf("Failed to consume WS ticket: %v", err)
		http.Error(w, "Invalid or expired ticket", http.StatusUnauthorized)
		return
	}

	// Upgrade connection using the new 2026 standard (coder/websocket)
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:3000", "app.synq.com"},
	})
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.CloseNow()

	// Inject claims into context
	ctx := r.Context()
	ctx = authcontext.WithTenantID(ctx, claims.TenantID)
	ctx = authcontext.WithOrgID(ctx, claims.OrgID)
	ctx = authcontext.WithUserID(ctx, claims.UserID)
	ctx = authcontext.WithRole(ctx, claims.Role)

	// Temporary dummy streaming loop for verification
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	_ = wsjson.Write(ctx, conn, AGUIIntent{
		Type: "connection_established",
		Data: map[string]string{"tenant_id": claims.TenantID, "role": claims.Role},
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := wsjson.Write(ctx, conn, AGUIIntent{
				Type:      "ui_update",
				Component: "progress",
				Data: map[string]interface{}{
					"status":  "processing",
					"percent": 50,
				},
			})
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}
