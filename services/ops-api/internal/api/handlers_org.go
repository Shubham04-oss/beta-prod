package api

import (
	"encoding/json"
	"fmt"
	"github.com/synq/pkg/authcontext"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synq/ops-api/internal/service"
	"github.com/synq/pkg/db"
)

type OrganizationHandlers struct {
	q                *db.Queries
	lifecycleService *service.LifecycleService
}

func NewOrganizationHandlers(q *db.Queries, lifecycleService *service.LifecycleService) *OrganizationHandlers {
	return &OrganizationHandlers{
		q:                q,
		lifecycleService: lifecycleService,
	}
}

func (h *OrganizationHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/organization/members", h.ListMembersHandler)
	r.Post("/api/v1/organization/members", h.InviteMemberHandler)
}

func (h *OrganizationHandlers) ListMembersHandler(w http.ResponseWriter, r *http.Request) {
	tenantIDStr, err := authcontext.GetTenantID(r.Context())
	if err != nil || tenantIDStr == "" {
		http.Error(w, "Unauthorized: Missing Tenant ID", http.StatusForbidden)
		return
	}

	var tenantUUID pgtype.UUID
	err = tenantUUID.Scan(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid Tenant ID", http.StatusBadRequest)
		return
	}

	members, err := h.q.ListTenantMembers(r.Context(), tenantUUID)
	if err != nil {
		log.Printf("Failed to list members: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *OrganizationHandlers) InviteMemberHandler(w http.ResponseWriter, r *http.Request) {
	role, err := authcontext.GetRole(r.Context())
	if err != nil || strings.ToUpper(role) != "ADMIN" {
		http.Error(w, "Forbidden: Only Administrators can invite members", http.StatusForbidden)
		return
	}

	tenantIDStr, _ := authcontext.GetTenantID(r.Context())
	orgIDStr, _ := authcontext.GetOrgID(r.Context())

	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		http.Error(w, "Role cannot be empty", http.StatusBadRequest)
		return
	}

	userRecord, err := h.lifecycleService.InviteUser(r.Context(), req.Email, orgIDStr, tenantIDStr, req.Role)
	if err != nil {
		fmt.Printf("Internal Error provisioning user: %v\n", err)
		http.Error(w, "An unexpected error occurred while provisioning the user account. Please contact support if the issue persists.", http.StatusInternalServerError)
		return
	}

	var orgUUID, tenantUUID pgtype.UUID
	orgUUID.Scan(orgIDStr)
	tenantUUID.Scan(tenantIDStr)

	detailsJSON, _ := json.Marshal(map[string]string{"email": req.Email, "role": req.Role})

	_, _ = h.q.InsertAuditEvent(r.Context(), db.InsertAuditEventParams{
		OrgID:      orgUUID,
		TenantID:   tenantUUID,
		ActorEmail: pgtype.Text{String: "admin@synq.app", Valid: true},
		Action:     "USER_INVITED",
		EntityType: "USER",
		EntityID:   pgtype.UUID{},
		Details:    detailsJSON,
		IpAddress:  pgtype.Text{String: r.RemoteAddr, Valid: true},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User successfully invited",
		"uid":     userRecord.UID,
	})
}
