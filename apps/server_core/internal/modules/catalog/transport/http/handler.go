package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"metalshopping/server_core/internal/modules/catalog/application"
	"metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/modules/catalog/ports"
	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type Handler struct {
	service           *application.Service
	permissionChecker ports.PermissionChecker
}

type CreateProductRequest struct {
	SKU    string `json:"sku"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

func NewHandler(service *application.Service, permissionChecker ports.PermissionChecker) *Handler {
	return &Handler{
		service:           service,
		permissionChecker: permissionChecker,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/catalog/products", h.handleProducts)
}

func (h *Handler) handleProducts(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.handleCreateProduct(w, r, principal.SubjectID, tenant.ID)
	case http.MethodGet:
		h.handleListProducts(w, r, principal.SubjectID, tenant.ID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleCreateProduct(w http.ResponseWriter, r *http.Request, userID, tenantID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermCatalogWrite)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", requestTraceID(r))
		return
	}

	product, err := h.service.CreateProduct(r.Context(), application.CreateProductCommand{
		TenantID: tenantID,
		SKU:      req.SKU,
		Name:     req.Name,
		Status:   req.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTenantIDRequired), errors.Is(err, domain.ErrSKURequired), errors.Is(err, domain.ErrProductNameRequired), errors.Is(err, domain.ErrInvalidProductStatus):
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
		case errors.Is(err, domain.ErrProductCreationDisabled):
			writeAPIError(w, http.StatusForbidden, "GOVERNANCE_DISABLED", err.Error(), requestTraceID(r))
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create product", requestTraceID(r))
		}
		return
	}

	writeJSON(w, http.StatusCreated, mapProduct(product))
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request, userID, tenantID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermCatalogRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	products, err := h.service.ListProducts(r.Context(), tenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list products", requestTraceID(r))
		return
	}

	items := make([]map[string]any, 0, len(products))
	for _, product := range products {
		items = append(items, mapProduct(product))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func mapProduct(product domain.Product) map[string]any {
	return map[string]any{
		"product_id": product.ProductID,
		"tenant_id":  product.TenantID,
		"sku":        product.SKU,
		"name":       product.Name,
		"status":     string(product.Status),
	}
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
	TraceID string         `json:"trace_id"`
}

func requestTraceID(r *http.Request) string {
	if traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id")); traceID != "" {
		return traceID
	}
	return "trace-local"
}

func writeAPIError(w http.ResponseWriter, status int, code, message, traceID string) {
	writeJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: map[string]any{},
			TraceID: traceID,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
