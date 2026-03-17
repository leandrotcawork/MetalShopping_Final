package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	"metalshopping/server_core/internal/modules/pricing/application"
	"metalshopping/server_core/internal/modules/pricing/domain"
	"metalshopping/server_core/internal/modules/pricing/ports"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type Handler struct {
	service           *application.Service
	permissionChecker ports.PermissionChecker
}

type SetProductPriceRequest struct {
	CurrencyCode     string  `json:"currency_code"`
	PriceAmount      float64 `json:"price_amount"`
	CostBasisAmount  float64 `json:"cost_basis_amount"`
	MarginFloorValue float64 `json:"margin_floor_value,omitempty"`
	PricingStatus    string  `json:"pricing_status,omitempty"`
	EffectiveFrom    string  `json:"effective_from"`
	EffectiveTo      string  `json:"effective_to,omitempty"`
	OriginType       string  `json:"origin_type,omitempty"`
	OriginRef        string  `json:"origin_ref,omitempty"`
	ReasonCode       string  `json:"reason_code"`
}

func NewHandler(service *application.Service, permissionChecker ports.PermissionChecker) *Handler {
	return &Handler{
		service:           service,
		permissionChecker: permissionChecker,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/pricing/products/", h.handleProductPrices)
}

func (h *Handler) handleProductPrices(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/pricing/products/")
	productID, subresource, found := strings.Cut(path, "/")
	if !found || strings.TrimSpace(productID) == "" {
		http.NotFound(w, r)
		return
	}
	productID = strings.TrimSpace(productID)

	switch {
	case subresource == "prices" && r.Method == http.MethodPost:
		h.handleSetProductPrice(w, r, principal.SubjectID, tenant.ID, productID)
	case subresource == "prices" && r.Method == http.MethodGet:
		h.handleListProductPrices(w, r, principal.SubjectID, tenant.ID, productID)
	case subresource == "prices/current" && r.Method == http.MethodGet:
		h.handleGetCurrentProductPrice(w, r, principal.SubjectID, tenant.ID, productID)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handleSetProductPrice(w http.ResponseWriter, r *http.Request, userID, tenantID, productID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermPricingWrite)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	var req SetProductPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", requestTraceID(r))
		return
	}

	effectiveFrom, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EffectiveFrom))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "effective_from must be a valid RFC3339 timestamp", requestTraceID(r))
		return
	}
	var effectiveTo *time.Time
	if strings.TrimSpace(req.EffectiveTo) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EffectiveTo))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "effective_to must be a valid RFC3339 timestamp", requestTraceID(r))
			return
		}
		effectiveTo = &parsed
	}

	price, err := h.service.SetProductPrice(r.Context(), application.SetProductPriceCommand{
		TenantID:         tenantID,
		TraceID:          requestTraceID(r),
		ProductID:        productID,
		CurrencyCode:     req.CurrencyCode,
		PriceAmount:      req.PriceAmount,
		CostBasisAmount:  req.CostBasisAmount,
		MarginFloorValue: req.MarginFloorValue,
		PricingStatus:    req.PricingStatus,
		EffectiveFrom:    effectiveFrom,
		EffectiveTo:      effectiveTo,
		OriginType:       req.OriginType,
		OriginRef:        req.OriginRef,
		ReasonCode:       req.ReasonCode,
		UpdatedBy:        userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTenantIDRequired),
			errors.Is(err, domain.ErrProductIDRequired),
			errors.Is(err, domain.ErrCurrencyCodeRequired),
			errors.Is(err, domain.ErrInvalidCurrencyCode),
			errors.Is(err, domain.ErrPriceAmountInvalid),
			errors.Is(err, domain.ErrCostBasisAmountInvalid),
			errors.Is(err, domain.ErrMarginFloorValueInvalid),
			errors.Is(err, domain.ErrInvalidPricingStatus),
			errors.Is(err, domain.ErrInvalidOriginType),
			errors.Is(err, domain.ErrReasonCodeRequired),
			errors.Is(err, domain.ErrUpdatedByRequired),
			errors.Is(err, domain.ErrEffectiveFromRequired),
			errors.Is(err, domain.ErrInvalidEffectiveWindow):
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
		case errors.Is(err, domain.ErrManualPriceOverrideDisabled), errors.Is(err, domain.ErrPriceBelowMarginFloor):
			writeAPIError(w, http.StatusForbidden, "GOVERNANCE_DISABLED", err.Error(), requestTraceID(r))
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set product price", requestTraceID(r))
		}
		return
	}

	writeJSON(w, http.StatusCreated, mapProductPrice(price))
}

func (h *Handler) handleListProductPrices(w http.ResponseWriter, r *http.Request, userID, tenantID, productID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermPricingRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	items, err := h.service.ListProductPrices(r.Context(), tenantID, productID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list product prices", requestTraceID(r))
		return
	}

	payload := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload = append(payload, mapProductPrice(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": payload})
}

func (h *Handler) handleGetCurrentProductPrice(w http.ResponseWriter, r *http.Request, userID, tenantID, productID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermPricingRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	item, err := h.service.GetCurrentProductPrice(r.Context(), tenantID, productID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductPriceNotFound):
			writeAPIError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), requestTraceID(r))
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load current product price", requestTraceID(r))
		}
		return
	}
	writeJSON(w, http.StatusOK, mapProductPrice(item))
}

func (h *Handler) requirePrincipalAndTenant(w http.ResponseWriter, r *http.Request) (platformauth.Principal, tenancy_runtime.Tenant, bool) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return platformauth.Principal{}, tenancy_runtime.Tenant{}, false
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return platformauth.Principal{}, tenancy_runtime.Tenant{}, false
	}
	return principal, tenant, true
}

func mapProductPrice(price domain.ProductPrice) map[string]any {
	payload := map[string]any{
		"price_id":           price.PriceID,
		"tenant_id":          price.TenantID,
		"product_id":         price.ProductID,
		"currency_code":      price.CurrencyCode,
		"price_amount":       price.PriceAmount,
		"cost_basis_amount":  price.CostBasisAmount,
		"margin_floor_value": price.MarginFloorValue,
		"pricing_status":     string(price.PricingStatus),
		"effective_from":     price.EffectiveFrom.Format(time.RFC3339),
		"origin_type":        string(price.OriginType),
		"origin_ref":         price.OriginRef,
		"reason_code":        price.ReasonCode,
		"updated_by":         price.UpdatedBy,
	}
	if price.EffectiveTo != nil {
		payload["effective_to"] = price.EffectiveTo.Format(time.RFC3339)
	}
	return payload
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
