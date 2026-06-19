package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/synq/ops-api/internal/service"
)

type TenantHandler struct {
	lifecycleService *service.LifecycleService
}

func NewTenantHandler(lifecycleService *service.LifecycleService) *TenantHandler {
	return &TenantHandler{
		lifecycleService: lifecycleService,
	}
}

func (h *TenantHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/onboard", h.HandleOnboardTenant)
}

type OnboardRequest struct {
	OrgName       string `json:"org_name"`
	TenantName    string `json:"tenant_name"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
}

func (h *TenantHandler) HandleOnboardTenant(w http.ResponseWriter, r *http.Request) {
	// TODO: Add strict IP-based rate limiting here to prevent tenant creation flooding
	var req OnboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OrgName == "" || req.TenantName == "" || req.AdminEmail == "" || req.AdminPassword == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.lifecycleService.CreateTenantLifecycle(r.Context(), req.OrgName, req.TenantName, req.AdminEmail, req.AdminPassword)
	if err != nil {
		log.Printf("Failed to create tenant: %v", err)
		http.Error(w, "Failed to provision tenant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success","message":"Tenant and Admin User created successfully"}`))
}
