package inventory

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type StockRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type StockResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func getTenancy(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	tenantID, err := uuid.Parse(r.Header.Get("X-Tenant-ID"))
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	orgID, err := uuid.Parse(r.Header.Get("X-Org-ID"))
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return tenantID, orgID, nil
}

func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	var req StockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ProductID == "" {
		http.Error(w, "missing product_id", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(req.ProductID); err != nil {
		http.Error(w, "invalid product_id format", http.StatusBadRequest)
		return
	}

	if req.Quantity == 0 {
		http.Error(w, "invalid quantity", http.StatusBadRequest)
		return
	}

	tenantID, orgID, tenancyErr := getTenancy(r)
	if tenancyErr != nil {
		http.Error(w, "invalid or missing tenancy headers", http.StatusBadRequest)
		return
	}

	err := h.svc.AdjustStock(r.Context(), tenantID, orgID, req.ProductID, req.Quantity)
	if err != nil {
		if errorsIs(err, ErrInvalidQuantity) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errorsIs(err, ErrInsufficientStock) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	var req StockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ProductID == "" {
		http.Error(w, "missing product_id", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(req.ProductID); err != nil {
		http.Error(w, "invalid product_id format", http.StatusBadRequest)
		return
	}

	tenantID, orgID, tenancyErr := getTenancy(r)
	if tenancyErr != nil {
		http.Error(w, "invalid or missing tenancy headers", http.StatusBadRequest)
		return
	}

	err := h.svc.ReserveStock(r.Context(), tenantID, orgID, req.ProductID, req.Quantity)
	if err != nil {
		if errorsIs(err, ErrInvalidQuantity) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errorsIs(err, ErrInsufficientStock) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ReleaseStock(w http.ResponseWriter, r *http.Request) {
	var req StockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ProductID == "" {
		http.Error(w, "missing product_id", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(req.ProductID); err != nil {
		http.Error(w, "invalid product_id format", http.StatusBadRequest)
		return
	}

	tenantID, orgID, tenancyErr := getTenancy(r)
	if tenancyErr != nil {
		http.Error(w, "invalid or missing tenancy headers", http.StatusBadRequest)
		return
	}

	err := h.svc.ReleaseStock(r.Context(), tenantID, orgID, req.ProductID, req.Quantity)
	if err != nil {
		if errorsIs(err, ErrInvalidQuantity) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errorsIs(err, ErrInsufficientStock) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetStock(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		http.Error(w, "missing product_id", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(productID); err != nil {
		http.Error(w, "invalid product_id format", http.StatusBadRequest)
		return
	}

	tenantID, orgID, tenancyErr := getTenancy(r)
	if tenancyErr != nil {
		http.Error(w, "invalid or missing tenancy headers", http.StatusBadRequest)
		return
	}

	avail, err := h.svc.GetStock(r.Context(), tenantID, orgID, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := StockResponse{
		ProductID: productID,
		Quantity:  avail,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func errorsIs(err, target error) bool {
	return err != nil && err.Error() == target.Error()
}

func RegisterRoutes(mux *http.ServeMux, svc Service) {
	h := NewHandler(svc)
	mux.HandleFunc("POST /api/inventory/adjust", h.AdjustStock)
	mux.HandleFunc("POST /api/inventory/reserve", h.ReserveStock)
	mux.HandleFunc("POST /api/inventory/release", h.ReleaseStock)
	mux.HandleFunc("GET /api/inventory/stock/{id}", h.GetStock)
}
