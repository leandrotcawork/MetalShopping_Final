package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/shopping/adapters/postgres"
	"metalshopping/server_core/internal/modules/shopping/application"
	"metalshopping/server_core/internal/modules/shopping/ports"
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
	mux.HandleFunc("/api/v1/shopping/summary", h.handleSummary)
	mux.HandleFunc("/api/v1/shopping/runs", h.handleRunsList)
	mux.HandleFunc("/api/v1/shopping/runs/", h.handleRunByID)
	mux.HandleFunc("/api/v1/shopping/products/", h.handleProductLatest)
}

func (h *Handler) handleSummary(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_summary", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	summary, err := h.service.GetSummary(r.Context(), tenantID)
	if err != nil {
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping summary", traceID)
		return
	}

	lastRunAt := time.Now().UTC()
	if summary.LastRunAt != nil {
		lastRunAt = summary.LastRunAt.UTC()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"totalRuns":     summary.TotalRuns,
		"runningRuns":   summary.RunningRuns,
		"completedRuns": summary.CompletedRuns,
		"failedRuns":    summary.FailedRuns,
		"lastRunAt":     lastRunAt.Format(time.RFC3339),
	})
}

func (h *Handler) handleRunsList(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.list_runs", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	limit := parseQueryInt64(r, "limit", 50)
	offset := parseQueryInt64(r, "offset", 0)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	runList, err := h.service.ListRuns(r.Context(), tenantID, ports.RunListFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list shopping runs", traceID)
		return
	}

	rows := make([]map[string]any, 0, len(runList.Rows))
	for _, run := range runList.Rows {
		rows = append(rows, mapRun(run))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rows": rows,
		"paging": map[string]any{
			"offset":   runList.Offset,
			"limit":    runList.Limit,
			"returned": len(runList.Rows),
			"total":    runList.Total,
		},
	})
}

func (h *Handler) handleRunByID(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_run", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	runID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/v1/shopping/runs/"))
	if runID == "" || strings.Contains(runID, "/") {
		statusCode = http.StatusNotFound
		reqResult = "not_found"
		writeShoppingError(w, http.StatusNotFound, "SHOPPING_RUN_NOT_FOUND", "Shopping run not found", traceID)
		return
	}

	run, err := h.service.GetRun(r.Context(), tenantID, runID)
	if err != nil {
		if errors.Is(err, postgres.ErrRunNotFound) {
			statusCode = http.StatusNotFound
			reqResult = "not_found"
			writeShoppingError(w, http.StatusNotFound, "SHOPPING_RUN_NOT_FOUND", "Shopping run not found", traceID)
			return
		}
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping run", traceID)
		return
	}

	writeJSON(w, http.StatusOK, mapRun(run))
}

func (h *Handler) handleProductLatest(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("shopping.get_product_latest", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = http.StatusUnauthorized
		reqResult = "auth_or_tenant_error"
		return
	}

	productID, routeOK := extractProductLatestPathParam(r.URL.Path)
	if !routeOK {
		statusCode = http.StatusNotFound
		reqResult = "not_found"
		writeShoppingError(w, http.StatusNotFound, "SHOPPING_PRODUCT_LATEST_NOT_FOUND", "Shopping latest snapshot not found", traceID)
		return
	}

	item, err := h.service.GetProductLatest(r.Context(), tenantID, productID)
	if err != nil {
		if errors.Is(err, postgres.ErrProductLatestNotFound) {
			statusCode = http.StatusNotFound
			reqResult = "not_found"
			writeShoppingError(w, http.StatusNotFound, "SHOPPING_PRODUCT_LATEST_NOT_FOUND", "Shopping latest snapshot not found", traceID)
			return
		}
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeShoppingError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load shopping latest snapshot", traceID)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"productId":     item.ProductID,
		"runId":         item.RunID,
		"observedAt":    item.ObservedAt.Format(time.RFC3339),
		"sellerName":    item.SellerName,
		"channel":       item.Channel,
		"observedPrice": item.ObservedPrice,
		"currency":      item.Currency,
	})
}

func mapRun(run ports.Run) map[string]any {
	var finishedAt any
	if run.FinishedAt != nil {
		finishedAt = run.FinishedAt.Format(time.RFC3339)
	}
	return map[string]any{
		"runId":          run.RunID,
		"status":         run.Status,
		"startedAt":      run.StartedAt.Format(time.RFC3339),
		"finishedAt":     finishedAt,
		"processedItems": run.ProcessedItems,
		"totalItems":     run.TotalItems,
		"notes":          run.Notes,
	}
}

func extractProductLatestPathParam(path string) (string, bool) {
	prefix := "/api/v1/shopping/products/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	suffix := strings.TrimPrefix(path, prefix)
	if !strings.HasSuffix(suffix, "/latest") {
		return "", false
	}
	productID := strings.TrimSuffix(suffix, "/latest")
	productID = strings.TrimSpace(strings.TrimSuffix(productID, "/"))
	if productID == "" || strings.Contains(productID, "/") {
		return "", false
	}
	return productID, true
}

func parseQueryInt64(r *http.Request, key string, fallback int64) int64 {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func authenticatedTenantID(w http.ResponseWriter, r *http.Request) (string, bool) {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeShoppingError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return "", false
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeShoppingError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return "", false
	}
	return tenant.ID, true
}

type shoppingErrorEnvelope struct {
	Error shoppingError `json:"error"`
}

type shoppingError struct {
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

func writeShoppingError(w http.ResponseWriter, status int, code, message, traceID string) {
	writeJSON(w, status, shoppingErrorEnvelope{
		Error: shoppingError{
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

func logRequest(action, traceID string, statusCode *int, result *string, startedAt time.Time) {
	durationMs := time.Since(startedAt).Milliseconds()
	slog.Info("shopping_request",
		"action", action,
		"trace_id", traceID,
		"result", *result,
		"status", *statusCode,
		"duration_ms", durationMs,
	)
}
