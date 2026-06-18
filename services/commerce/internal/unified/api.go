package unified

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type API struct {
	svc *SyncService
}

func NewAPI(svc *SyncService) *API {
	return &API{svc: svc}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/unified/sync/push/product", a.handlePushProduct)
	mux.HandleFunc("POST /api/unified/sync/pull/order", a.handlePullOrder)
	mux.HandleFunc("POST /api/unified/webhook", a.handleWebhook)
	mux.HandleFunc("GET /api/unified/sync/status/{id}", a.handleSyncStatus)
}

func (a *API) handlePushProduct(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProductID string `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.ProductID == "" {
		http.Error(w, "empty product_id", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(r.Header.Get("X-Tenant-ID"))
	if err != nil {
		http.Error(w, "invalid or missing X-Tenant-ID", http.StatusBadRequest)
		return
	}
	orgID, err := uuid.Parse(r.Header.Get("X-Org-ID"))
	if err != nil {
		http.Error(w, "invalid or missing X-Org-ID", http.StatusBadRequest)
		return
	}

	jobID := a.svc.GenerateJobID()
	a.svc.SetSyncStatus(jobID, "pending")

	err = a.svc.PushProduct(r.Context(), tenantID, orgID, "default-connection", payload.ProductID)
	if err != nil {
		a.svc.SetSyncStatus(jobID, "failed")
		if errors.Is(err, ErrProductNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	a.svc.SetSyncStatus(jobID, "completed")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"job_id": jobID,
	})
}

func (a *API) handlePullOrder(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OrderID string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.OrderID == "" {
		http.Error(w, "empty order_id", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(r.Header.Get("X-Tenant-ID"))
	if err != nil {
		http.Error(w, "invalid or missing X-Tenant-ID", http.StatusBadRequest)
		return
	}
	orgID, err := uuid.Parse(r.Header.Get("X-Org-ID"))
	if err != nil {
		http.Error(w, "invalid or missing X-Org-ID", http.StatusBadRequest)
		return
	}

	jobID := a.svc.GenerateJobID()
	a.svc.SetSyncStatus(jobID, "pending")

	order, err := a.svc.PullOrder(r.Context(), tenantID, orgID, "default-connection", payload.OrderID)
	if err != nil {
		a.svc.SetSyncStatus(jobID, "failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.svc.SetSyncStatus(jobID, "completed")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "success",
		"job_id":         jobID,
		"local_order_id": order.ID.String(),
	})
}

func (a *API) handleWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	var rawPayload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("X-Signature") == "" {
		http.Error(w, "missing signature", http.StatusUnauthorized)
		return
	}

	if len(rawPayload) == 0 {
		http.Error(w, "empty payload", http.StatusBadRequest)
		return
	}

	event, _ := rawPayload["event"].(string)
	if event != "inventory.updated" {
		http.Error(w, "unknown event", http.StatusBadRequest)
		return
	}

	tenantIDStr := r.Header.Get("X-Tenant-ID")
	orgIDStr := r.Header.Get("X-Org-ID")

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		tenantID = uuid.Nil
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = uuid.Nil
	}

	ctx := context.WithValue(r.Context(), "tenantID", tenantID)
	ctx = context.WithValue(ctx, "orgID", orgID)

	err = a.svc.HandleWebhook(ctx, event, rawPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
	})
}

func (a *API) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if jobID == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}

	status := a.svc.GetSyncStatus(jobID)
	if status == "not_found" {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
		"job_id": jobID,
	})
}
