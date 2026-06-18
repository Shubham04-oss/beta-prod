package oms

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"commerce_modules/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type API struct {
	svc *OMSService
}

func NewAPI(svc *OMSService) *API {
	return &API{svc: svc}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/oms/orders", a.handleCreateOrder)
	mux.HandleFunc("GET /api/oms/orders/{id}", a.handleGetOrder)
	mux.HandleFunc("PUT /api/oms/orders/{id}/status", a.handleUpdateOrderStatus)
	mux.HandleFunc("POST /api/oms/customers", a.handleCreateCustomer)
}

func extractTenantOrg(r *http.Request) (uuid.UUID, uuid.UUID) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	orgIDStr := r.Header.Get("X-Org-ID")

	tID, err := uuid.Parse(tenantIDStr)
	if err != nil && tenantIDStr != "" {
		tID = uuid.NewMD5(uuid.NameSpaceOID, []byte(tenantIDStr))
	} else if err != nil {
		tID = uuid.Nil
	}

	oID, err := uuid.Parse(orgIDStr)
	if err != nil && orgIDStr != "" {
		oID = uuid.NewMD5(uuid.NameSpaceOID, []byte(orgIDStr))
	} else if err != nil {
		oID = uuid.Nil
	}

	return tID, oID
}

func (a *API) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	tID, oID := extractTenantOrg(r)

	var req struct {
		CustomerID *string `json:"customer_id"`
		Items      []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "empty items", http.StatusBadRequest)
		return
	}
	if len(req.Items) > 1000 {
		http.Error(w, "too many items", http.StatusRequestEntityTooLarge)
		return
	}

	var items []models.OrderLineItem
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			http.Error(w, "invalid quantity", http.StatusBadRequest)
			return
		}

		var variantID uuid.UUID
		vID, err := uuid.Parse(item.ProductID)
		if err != nil {
			variantID = uuid.NewMD5(uuid.NameSpaceOID, []byte(item.ProductID))
		} else {
			variantID = vID
		}

		v := variantID // need address

		items = append(items, models.OrderLineItem{
			ID:        uuid.New(),
			TenantID:  tID,
			OrgID:     oID,
			VariantID: &v,
			Quantity:  item.Quantity,
		})
	}

	order := &models.Order{
		ID:       uuid.New(),
		TenantID: tID,
		OrgID:    oID,
	}

	if req.CustomerID != nil {
		cID, err := uuid.Parse(*req.CustomerID)
		if err != nil {
			cID = uuid.NewMD5(uuid.NameSpaceOID, []byte(*req.CustomerID))
		}
		order.CustomerID = &cID
	}

	err := a.svc.CreateOrder(r.Context(), tID, oID, order, items)
	if err != nil {
		log.Printf("OMS CreateOrder error: %v", err)
		if strings.Contains(err.Error(), "insufficient stock") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "variant not found") || strings.Contains(err.Error(), "product not found") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (a *API) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	tID, oID := extractTenantOrg(r)
	idStr := r.PathValue("id")

	orderID, err := uuid.Parse(idStr)
	if err != nil {
		orderID = uuid.NewMD5(uuid.NameSpaceOID, []byte(idStr))
	}

	order, err := a.svc.GetOrder(r.Context(), tID, oID, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(order)
}

func (a *API) handleUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	tID, oID := extractTenantOrg(r)
	idStr := r.PathValue("id")

	orderID, err := uuid.Parse(idStr)
	if err != nil {
		orderID = uuid.NewMD5(uuid.NameSpaceOID, []byte(idStr))
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status := strings.ToUpper(req["status"])
	if status == "CANCELLED" || status == "CANCELED" {
		err = a.svc.CancelOrder(r.Context(), tID, oID, orderID)
	} else if status == "FULFILLED" || status == "SHIPPED" {
		err = a.svc.FulfillOrder(r.Context(), tID, oID, orderID)
	} else {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if err != nil {
		if strings.Contains(err.Error(), "invalid state") || strings.Contains(err.Error(), "no rows") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedOrder, err := a.svc.GetOrder(r.Context(), tID, oID, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedOrder)
}

func (a *API) handleCreateCustomer(w http.ResponseWriter, r *http.Request) {
	tID, oID := extractTenantOrg(r)

	var cust models.Customer
	if err := json.NewDecoder(r.Body).Decode(&cust); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := a.svc.CreateCustomer(r.Context(), tID, oID, &cust)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cust)
}
