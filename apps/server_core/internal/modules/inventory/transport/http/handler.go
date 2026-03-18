package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	"metalshopping/server_core/internal/modules/inventory/application"
	"metalshopping/server_core/internal/modules/inventory/domain"
	"metalshopping/server_core/internal/modules/inventory/ports"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type Handler struct {
	service           *application.Service
	permissionChecker ports.PermissionChecker
}

type SetProductPositionRequest struct {
	OnHandQuantity float64 `json:"on_hand_quantity"`
	LastPurchaseAt string  `json:"last_purchase_at,omitempty"`
	LastSaleAt     string  `json:"last_sale_at,omitempty"`
	PositionStatus string  `json:"position_status,omitempty"`
	EffectiveFrom  string  `json:"effective_from"`
	EffectiveTo    string  `json:"effective_to,omitempty"`
	OriginType     string  `json:"origin_type,omitempty"`
	OriginRef      string  `json:"origin_ref,omitempty"`
	ReasonCode     string  `json:"reason_code"`
}

func NewHandler(service *application.Service, permissionChecker ports.PermissionChecker) *Handler {
	return &Handler{service: service, permissionChecker: permissionChecker}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/inventory/products/", h.handleProductPositions)
}

func (h *Handler) handleProductPositions(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/inventory/products/")
	productID, subresource, found := strings.Cut(path, "/")
	if !found || strings.TrimSpace(productID) == "" {
		http.NotFound(w, r)
		return
	}
	productID = strings.TrimSpace(productID)

	switch {
	case subresource == "positions" && r.Method == http.MethodPost:
		h.handleSetProductPosition(w, r, principal.SubjectID, tenant.ID, productID)
	case subresource == "positions" && r.Method == http.MethodGet:
		h.handleListProductPositions(w, r, principal.SubjectID, tenant.ID, productID)
	case subresource == "positions/current" && r.Method == http.MethodGet:
		h.handleGetCurrentProductPosition(w, r, principal.SubjectID, tenant.ID, productID)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handleSetProductPosition(w http.ResponseWriter, r *http.Request, userID, tenantID, productID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermInventoryWrite)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	var req SetProductPositionRequest
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
	lastPurchaseAt, err := parseOptionalTime(req.LastPurchaseAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "last_purchase_at must be a valid RFC3339 timestamp", requestTraceID(r))
		return
	}
	lastSaleAt, err := parseOptionalTime(req.LastSaleAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "last_sale_at must be a valid RFC3339 timestamp", requestTraceID(r))
		return
	}

	position, applied, err := h.service.SetProductPosition(r.Context(), application.SetProductPositionCommand{
		TenantID:       tenantID,
		TraceID:        requestTraceID(r),
		ProductID:      productID,
		OnHandQuantity: req.OnHandQuantity,
		LastPurchaseAt: lastPurchaseAt,
		LastSaleAt:     lastSaleAt,
		PositionStatus: req.PositionStatus,
		EffectiveFrom:  effectiveFrom,
		EffectiveTo:    effectiveTo,
		OriginType:     req.OriginType,
		OriginRef:      req.OriginRef,
		ReasonCode:     req.ReasonCode,
		UpdatedBy:      userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPositionIDRequired),
			errors.Is(err, domain.ErrTenantIDRequired),
			errors.Is(err, domain.ErrProductIDRequired),
			errors.Is(err, domain.ErrOnHandQuantityInvalid),
			errors.Is(err, domain.ErrInvalidPositionStatus),
			errors.Is(err, domain.ErrEffectiveFromRequired),
			errors.Is(err, domain.ErrInvalidEffectiveWindow),
			errors.Is(err, domain.ErrInvalidOriginType),
			errors.Is(err, domain.ErrReasonCodeRequired),
			errors.Is(err, domain.ErrUpdatedByRequired):
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set product position", requestTraceID(r))
		}
		return
	}

	w.Header().Set("X-Change-Applied", boolToString(applied))
	if applied {
		writeJSON(w, http.StatusCreated, mapProductPosition(position))
		return
	}
	writeJSON(w, http.StatusOK, mapProductPosition(position))
}

func (h *Handler) handleListProductPositions(w http.ResponseWriter, r *http.Request, userID, tenantID, productID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermInventoryRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	items, err := h.service.ListProductPositions(r.Context(), tenantID, productID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list product positions", requestTraceID(r))
		return
	}
	payload := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload = append(payload, mapProductPosition(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": payload})
}

func (h *Handler) handleGetCurrentProductPosition(w http.ResponseWriter, r *http.Request, userID, tenantID, productID string) {
	allowed, err := h.permissionChecker.HasPermission(r.Context(), userID, iamdomain.PermInventoryRead)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	item, err := h.service.GetCurrentProductPosition(r.Context(), tenantID, productID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductPositionNotFound):
			writeAPIError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), requestTraceID(r))
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load current product position", requestTraceID(r))
		}
		return
	}
	writeJSON(w, http.StatusOK, mapProductPosition(item))
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

func mapProductPosition(position domain.ProductPosition) map[string]any {
	payload := map[string]any{
		"position_id":      position.PositionID,
		"tenant_id":        position.TenantID,
		"product_id":       position.ProductID,
		"on_hand_quantity": position.OnHandQuantity,
		"position_status":  string(position.PositionStatus),
		"effective_from":   position.EffectiveFrom.Format(time.RFC3339),
		"origin_type":      string(position.OriginType),
		"origin_ref":       position.OriginRef,
		"reason_code":      position.ReasonCode,
		"updated_by":       position.UpdatedBy,
	}
	if position.LastPurchaseAt != nil {
		payload["last_purchase_at"] = position.LastPurchaseAt.Format(time.RFC3339)
	}
	if position.LastSaleAt != nil {
		payload["last_sale_at"] = position.LastSaleAt.Format(time.RFC3339)
	}
	if position.EffectiveTo != nil {
		payload["effective_to"] = position.EffectiveTo.Format(time.RFC3339)
	}
	return payload
}

func parseOptionalTime(raw string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
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
		Error: apiError{Code: code, Message: message, Details: map[string]any{}, TraceID: traceID},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func boolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
