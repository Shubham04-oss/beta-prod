package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/client"

	"github.com/synq/ops-api/internal/oms"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

type OMSHandler struct {
	temporalClient client.Client
	queries        *db.Queries
}

func NewOMSHandler(tc client.Client, queries ...*db.Queries) *OMSHandler {
	var q *db.Queries
	if len(queries) > 0 {
		q = queries[0]
	}
	return &OMSHandler{temporalClient: tc, queries: q}
}

func (h *OMSHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/oms", func(r chi.Router) {
		r.Get("/orders", h.ListOrders)
		r.Get("/orders/{orderID}", h.GetOrder)
		r.Post("/orders", h.CreateOrder)
	})
}

func (h *OMSHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req oms.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.IdempotencyKey == nil || *req.IdempotencyKey == "" {
		http.Error(w, "missing idempotency_key", http.StatusBadRequest)
		return
	}
	if req.PaymentProvider == nil || *req.PaymentProvider == "" || req.PaymentReference == nil || *req.PaymentReference == "" {
		http.Error(w, "payment_provider and payment_reference are required", http.StatusBadRequest)
		return
	}

	tenantID, err := authcontext.GetTenantID(r.Context())
	if err != nil {
		http.Error(w, "tenant id required", http.StatusUnauthorized)
		return
	}
	orgID, err := authcontext.GetOrgID(r.Context())
	if err != nil {
		http.Error(w, "org id required", http.StatusUnauthorized)
		return
	}
	userID, err := authcontext.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "user id required", http.StatusUnauthorized)
		return
	}
	role, err := authcontext.GetRole(r.Context())
	if err != nil {
		http.Error(w, "role required", http.StatusUnauthorized)
		return
	}

	req.TenantID = tenantID
	req.OrgID = orgID
	req.UserID = userID
	req.Role = role

	reservations := make([]oms.LineItemReservation, 0, len(req.Items))
	for _, item := range req.Items {
		if item.VariantID == nil || *item.VariantID == "" {
			continue
		}
		if item.LocationID == nil || *item.LocationID == "" {
			http.Error(w, "location_id is required for inventory-backed items", http.StatusBadRequest)
			return
		}
		reservations = append(reservations, oms.LineItemReservation{
			VariantID:  *item.VariantID,
			LocationID: *item.LocationID,
			Quantity:   item.Quantity,
		})
	}

	params := oms.OrderCreationParams{
		Request:      req,
		Reservations: reservations,
	}

	// Start Temporal Workflow
	options := client.StartWorkflowOptions{
		ID:        "oms-create-order-" + *req.IdempotencyKey,
		TaskQueue: "oms-task-queue-v3",
	}

	// PillarContextPropagator will inject context IDs into the workflow headers
	we, err := h.temporalClient.ExecuteWorkflow(r.Context(), options, oms.OrderCreationWorkflow, params)
	if err != nil {
		http.Error(w, "Failed to execute order creation workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"workflow_id": we.GetID(),
		"run_id":      we.GetRunID(),
	})
}

func (h *OMSHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "orders query service unavailable", http.StatusServiceUnavailable)
		return
	}

	tenantUUID, ok := tenantUUIDFromContext(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}

	limit := int32(50)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = int32(parsed)
	}

	offset := int32(0)
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = int32(parsed)
	}

	orders, err := h.queries.ListOrders(r.Context(), db.ListOrdersParams{
		TenantID: tenantUUID,
		OrgID:    orgUUID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		http.Error(w, "failed to list orders", http.StatusInternalServerError)
		return
	}
	if orders == nil {
		orders = []db.Order{}
	}

	out := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		out = append(out, makeOrderResponse(order))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": out,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *OMSHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "orders query service unavailable", http.StatusServiceUnavailable)
		return
	}

	tenantUUID, ok := tenantUUIDFromContext(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}

	orderID, err := uuid.Parse(chi.URLParam(r, "orderID"))
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	order, err := h.queries.GetOrder(r.Context(), db.GetOrderParams{
		ID:       pgtype.UUID{Bytes: orderID, Valid: true},
		TenantID: tenantUUID,
		OrgID:    orgUUID,
	})
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(makeOrderResponse(order))
}

type orderResponse struct {
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	Currency         string   `json:"currency"`
	Subtotal         float64  `json:"subtotal"`
	DiscountTotal    float64  `json:"discount_total"`
	ShippingTotal    float64  `json:"shipping_total"`
	TaxTotal         float64  `json:"tax_total"`
	Total            float64  `json:"total"`
	PaymentStatus    *string  `json:"payment_status"`
	PaymentProvider  *string  `json:"payment_provider"`
	PaymentReference *string  `json:"payment_reference"`
	Channel          *string  `json:"channel"`
	SourcePlatform   *string  `json:"source_platform"`
	IdempotencyKey   *string  `json:"idempotency_key"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
	Tags             []string `json:"tags"`
}

func makeOrderResponse(order db.Order) orderResponse {
	return orderResponse{
		ID:               uuidFromPG(order.ID),
		Status:           string(order.Status),
		Currency:         order.Currency,
		Subtotal:         numericToFloat(order.Subtotal),
		DiscountTotal:    numericToFloat(order.DiscountTotal),
		ShippingTotal:    numericToFloat(order.ShippingTotal),
		TaxTotal:         numericToFloat(order.TaxTotal),
		Total:            numericToFloat(order.Total),
		PaymentStatus:    textPtr(order.PaymentStatus),
		PaymentProvider:  textPtr(order.PaymentProvider),
		PaymentReference: textPtr(order.PaymentReference),
		Channel:          textPtr(order.Channel),
		SourcePlatform:   textPtr(order.SourcePlatform),
		IdempotencyKey:   textPtr(order.IdempotencyKey),
		CreatedAt:        timeString(order.CreatedAt),
		UpdatedAt:        timeString(order.UpdatedAt),
		Tags:             order.Tags,
	}
}

func tenantUUIDFromContext(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
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

func uuidFromPG(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}

func textPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func timeString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func numericToFloat(value pgtype.Numeric) float64 {
	f, err := value.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}
