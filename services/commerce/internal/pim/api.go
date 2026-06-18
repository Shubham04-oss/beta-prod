package pim

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/shopspring/decimal"

	"commerce_modules/internal/models"
)

type API struct {
	service *Service
}

func NewAPI(service *Service) *API {
	return &API{service: service}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/pim/products", a.handleCreateProduct)
	mux.HandleFunc("GET /api/pim/products/{id}", a.handleGetProduct)
	mux.HandleFunc("PUT /api/pim/products/{id}", a.handleUpdateProduct)
	mux.HandleFunc("DELETE /api/pim/products/{id}", a.handleDeleteProduct)
	mux.HandleFunc("POST /api/pim/products/search", a.handleSearchProducts)
	mux.HandleFunc("POST /api/pim/products/{id}/variants", a.handleCreateVariant)
	mux.HandleFunc("GET /api/pim/variants/{id}", a.handleGetVariant)
	mux.HandleFunc("GET /api/pim/products/{id}/variants", a.handleListVariants)
	mux.HandleFunc("PUT /api/pim/variants/{id}", a.handleUpdateVariant)
	mux.HandleFunc("DELETE /api/pim/variants/{id}", a.handleDeleteVariant)
}

func extractTenantInfo(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	orgIDStr := r.Header.Get("X-Org-ID")
	if orgIDStr == "" {
		return uuid.Nil, uuid.Nil, errors.New("missing X-Org-ID header")
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = uuid.NewMD5(uuid.NameSpaceOID, []byte(orgIDStr))
	}

	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return uuid.Nil, uuid.Nil, errors.New("missing X-Tenant-ID header")
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		tenantID = uuid.NewMD5(uuid.NameSpaceOID, []byte(tenantIDStr))
	}

	return orgID, tenantID, nil
}

func sendError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func (a *API) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	var req struct {
		SKU         string          `json:"sku"`
		Name        string          `json:"name"`
		Description *string         `json:"description"`
		Price       decimal.Decimal `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}
	if req.SKU == "" {
		sendError(w, http.StatusBadRequest, errors.New("empty sku"))
		return
	}
	if req.Price.IsNegative() {
		sendError(w, http.StatusBadRequest, errors.New("negative price"))
		return
	}

	product := models.Product{
		ID:          uuid.New(),
		OrgID:       orgID,
		TenantID:    tenantID,
		Title:       req.Name,
		Description: req.Description,
		Status:      "ACTIVE",
	}

	if err := a.service.CreateProduct(r.Context(), &product); err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sku := req.SKU
	variant := models.ProductVariant{
		ID:        uuid.New(),
		OrgID:     orgID,
		TenantID:  tenantID,
		ProductID: product.ID,
		SKU:       &sku,
		Price:     req.Price,
		Currency:  "USD",
	}

	if err := a.service.CreateVariant(r.Context(), &variant); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			sendError(w, http.StatusConflict, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	// Just return product with its ID for tests
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    product.ID,
		"sku":   req.SKU,
		"name":  req.Name,
		"price": req.Price.InexactFloat64(),
	})
}

func (a *API) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	productID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusNotFound, errors.New("not found"))
		return
	}

	product, err := a.service.GetProduct(r.Context(), orgID, tenantID, productID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			sendError(w, http.StatusNotFound, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func (a *API) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	productID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusNotFound, errors.New("not found"))
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}
	if req.Name == "" {
		sendError(w, http.StatusBadRequest, errors.New("empty payload or title"))
		return
	}

	product := models.Product{
		ID:          productID,
		OrgID:       orgID,
		TenantID:    tenantID,
		Title:       req.Name,
		Description: req.Description,
	}

	if err := a.service.UpdateProduct(r.Context(), &product); err != nil {
		if errors.Is(err, ErrNotFound) {
			sendError(w, http.StatusNotFound, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func (a *API) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	productID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusNotFound, errors.New("not found"))
		return
	}

	if err := a.service.DeleteProduct(r.Context(), orgID, tenantID, productID); err != nil {
		if errors.Is(err, ErrNotFound) {
			sendError(w, http.StatusNotFound, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type SearchRequest struct {
	Embedding []float32 `json:"embedding"`
	Limit     int       `json:"limit"`
}

func (a *API) handleSearchProducts(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	vec := pgvector.NewVector(req.Embedding)

	products, err := a.service.SearchProducts(r.Context(), orgID, tenantID, vec, limit)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	if products == nil {
		products = []*models.Product{}
	}

	json.NewEncoder(w).Encode(products)
}

func (a *API) handleCreateVariant(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	productID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusBadRequest, errors.New("invalid product id"))
		return
	}

	var variant models.ProductVariant
	if err := json.NewDecoder(r.Body).Decode(&variant); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	variant.ID = uuid.New()
	variant.OrgID = orgID
	variant.TenantID = tenantID
	variant.ProductID = productID

	if err := a.service.CreateVariant(r.Context(), &variant); err != nil {
		if errors.Is(err, ErrDuplicateSKU) {
			sendError(w, http.StatusConflict, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(variant)
}

func (a *API) handleGetVariant(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	variantID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusBadRequest, errors.New("invalid variant id"))
		return
	}

	variant, err := a.service.GetVariant(r.Context(), orgID, tenantID, variantID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			sendError(w, http.StatusNotFound, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(variant)
}

func (a *API) handleListVariants(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	productID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusBadRequest, errors.New("invalid product id"))
		return
	}

	variants, err := a.service.ListVariants(r.Context(), orgID, tenantID, productID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	if variants == nil {
		variants = []*models.ProductVariant{}
	}

	json.NewEncoder(w).Encode(variants)
}

func (a *API) handleUpdateVariant(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	variantID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusBadRequest, errors.New("invalid variant id"))
		return
	}

	var variant models.ProductVariant
	if err := json.NewDecoder(r.Body).Decode(&variant); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	variant.ID = variantID
	variant.OrgID = orgID
	variant.TenantID = tenantID

	if err := a.service.UpdateVariant(r.Context(), &variant); err != nil {
		if errors.Is(err, ErrNotFound) {
			sendError(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, ErrDuplicateSKU) {
			sendError(w, http.StatusConflict, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(variant)
}

func (a *API) handleDeleteVariant(w http.ResponseWriter, r *http.Request) {
	orgID, tenantID, err := extractTenantInfo(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err)
		return
	}

	variantID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, http.StatusBadRequest, errors.New("invalid variant id"))
		return
	}

	if err := a.service.DeleteVariant(r.Context(), orgID, tenantID, variantID); err != nil {
		if errors.Is(err, ErrNotFound) {
			sendError(w, http.StatusNotFound, err)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
