package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type HomeHandler struct {
	db *sql.DB
}

func NewHomeHandler(db *sql.DB) *HomeHandler {
	return &HomeHandler{db: db}
}

func (h *HomeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/home/summary", h.handleHomeSummary)
}

func (h *HomeHandler) handleHomeSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeHomeError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return
	}
	_, ok = tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeHomeError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return
	}

	const query = `
SELECT
  (SELECT COUNT(*) FROM catalog_products WHERE tenant_id = current_tenant_id()) AS product_count,
  (SELECT COUNT(*) FROM catalog_products WHERE tenant_id = current_tenant_id() AND status = 'active') AS active_product_count,
  (SELECT COUNT(*) FROM pricing_product_prices WHERE tenant_id = current_tenant_id() AND effective_to IS NULL) AS priced_product_count,
  (SELECT COUNT(*) FROM inventory_product_positions WHERE tenant_id = current_tenant_id() AND effective_to IS NULL) AS inventory_tracked_count,
  (
    SELECT MAX(updated_at)
    FROM (
      SELECT MAX(updated_at) AS updated_at FROM catalog_products WHERE tenant_id = current_tenant_id()
      UNION ALL
      SELECT MAX(updated_at) AS updated_at FROM pricing_product_prices WHERE tenant_id = current_tenant_id()
      UNION ALL
      SELECT MAX(updated_at) AS updated_at FROM inventory_product_positions WHERE tenant_id = current_tenant_id()
    ) all_updates
  ) AS last_updated
`

	var productCount int64
	var activeProductCount int64
	var pricedProductCount int64
	var inventoryTrackedCount int64
	var lastUpdated sql.NullTime

	if err := h.db.QueryRowContext(r.Context(), query).Scan(
		&productCount,
		&activeProductCount,
		&pricedProductCount,
		&inventoryTrackedCount,
		&lastUpdated,
	); err != nil {
		writeHomeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load home summary", requestTraceID(r))
		return
	}

	updatedAt := time.Now().UTC()
	if lastUpdated.Valid {
		updatedAt = lastUpdated.Time.UTC()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"productCount":          productCount,
		"activeProductCount":    activeProductCount,
		"pricedProductCount":    pricedProductCount,
		"inventoryTrackedCount": inventoryTrackedCount,
		"lastUpdated":           updatedAt.Format(time.RFC3339),
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
