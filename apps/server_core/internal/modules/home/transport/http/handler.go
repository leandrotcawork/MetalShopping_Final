package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/home/application"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/home/summary", h.handleHomeSummary)
}

func (h *Handler) handleHomeSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeHomeError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeHomeError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return
	}

	summary, err := h.service.GetSummary(r.Context(), tenant.ID)
	if err != nil {
		writeHomeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load home summary", requestTraceID(r))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"productCount":          summary.ProductCount,
		"activeProductCount":    summary.ActiveProductCount,
		"pricedProductCount":    summary.PricedProductCount,
		"inventoryTrackedCount": summary.InventoryTrackedCount,
		"lastUpdated":           summary.LastUpdated.Format(time.RFC3339),
	})
}

type homeErrorEnvelope struct {
	Error homeError `json:"error"`
}

type homeError struct {
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

func writeHomeError(w http.ResponseWriter, status int, code, message, traceID string) {
	writeJSON(w, status, homeErrorEnvelope{
		Error: homeError{
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
