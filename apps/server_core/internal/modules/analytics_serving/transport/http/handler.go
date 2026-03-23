package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/analytics_serving/application"
	"metalshopping/server_core/internal/modules/analytics_serving/ports"
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
	mux.HandleFunc("/api/v1/analytics/home", h.handleGetHome)
}

func (h *Handler) handleGetHome(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	traceID := requestTraceID(r)
	statusCode := http.StatusOK
	reqResult := "success"
	defer logRequest("analytics.get_home", traceID, &statusCode, &reqResult, startedAt)

	if r.Method != http.MethodGet {
		statusCode = http.StatusMethodNotAllowed
		reqResult = "method_not_allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, authStatus, ok := authenticatedTenantID(w, r)
	if !ok {
		statusCode = authStatus
		reqResult = "auth_or_tenant_error"
		return
	}

	requestedSnapshotID := strings.TrimSpace(r.URL.Query().Get("snapshot_id"))
	home, err := h.service.GetHome(r.Context(), tenantID, requestedSnapshotID)
	if err != nil {
		statusCode = http.StatusInternalServerError
		reqResult = "internal_error"
		writeAnalyticsError(w, http.StatusInternalServerError, "ANALYTICS_HOME_INTERNAL_ERROR", "Failed to load analytics home", traceID)
		return
	}

	writeJSON(w, http.StatusOK, mapHomeResponse(home))
}

type homeResponse struct {
	SchemaVersion string             `json:"schemaVersion"`
	Snapshot      homeSnapshot       `json:"snapshot"`
	Blocks        homeBlocksResponse `json:"blocks"`
}

type homeSnapshot struct {
	RequestedID string  `json:"requestedId"`
	ResolvedID  *string `json:"resolvedId"`
	AsOf        *string `json:"asOf"`
	ServedAt    string  `json:"servedAt"`
}

type homeBlocksResponse struct {
	KpisOperational       homeBlockResponse `json:"kpisOperational"`
	KpisAnalytics         homeBlockResponse `json:"kpisAnalytics"`
	KpisProducts          homeBlockResponse `json:"kpisProducts"`
	ActionsToday          homeBlockResponse `json:"actionsToday"`
	AlertsPrioritarios    homeBlockResponse `json:"alertsPrioritarios"`
	PortfolioDistribution homeBlockResponse `json:"portfolioDistribution"`
	Timeline              homeBlockResponse `json:"timeline"`
}

type homeBlockResponse struct {
	Status string          `json:"status"`
	Data   map[string]any  `json:"data"`
	Error  *homeBlockError `json:"error"`
}

type homeBlockError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

func mapHomeResponse(home ports.Home) homeResponse {
	var asOf *string
	if home.Snapshot.AsOf != nil {
		value := home.Snapshot.AsOf.UTC().Format(time.RFC3339)
		asOf = &value
	}
	return homeResponse{
		SchemaVersion: home.SchemaVersion,
		Snapshot: homeSnapshot{
			RequestedID: home.Snapshot.RequestedID,
			ResolvedID:  home.Snapshot.ResolvedID,
			AsOf:        asOf,
			ServedAt:    home.Snapshot.ServedAt.UTC().Format(time.RFC3339),
		},
		Blocks: homeBlocksResponse{
			KpisOperational:       mapBlock(home.Blocks.KpisOperational),
			KpisAnalytics:         mapBlock(home.Blocks.KpisAnalytics),
			KpisProducts:          mapBlock(home.Blocks.KpisProducts),
			ActionsToday:          mapBlock(home.Blocks.ActionsToday),
			AlertsPrioritarios:    mapBlock(home.Blocks.AlertsPrioritarios),
			PortfolioDistribution: mapBlock(home.Blocks.PortfolioDistribution),
			Timeline:              mapBlock(home.Blocks.Timeline),
		},
	}
}

func mapBlock(block ports.HomeBlock) homeBlockResponse {
	var blockError *homeBlockError
	if block.Error != nil {
		blockError = &homeBlockError{
			Code:    block.Error.Code,
			Message: block.Error.Message,
			Details: block.Error.Details,
		}
	}
	return homeBlockResponse{
		Status: string(block.Status),
		Data:   block.Data,
		Error:  blockError,
	}
}

func authenticatedTenantID(w http.ResponseWriter, r *http.Request) (string, int, bool) {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeAnalyticsError(w, http.StatusUnauthorized, "ANALYTICS_HOME_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return "", http.StatusUnauthorized, false
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeAnalyticsError(w, http.StatusForbidden, "ANALYTICS_HOME_TENANT_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return "", http.StatusForbidden, false
	}
	return tenant.ID, http.StatusOK, true
}

type analyticsErrorEnvelope struct {
	Error analyticsError `json:"error"`
}

type analyticsError struct {
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

func writeAnalyticsError(w http.ResponseWriter, status int, code, message, traceID string) {
	writeJSON(w, status, analyticsErrorEnvelope{
		Error: analyticsError{
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
	slog.Info("analytics_request",
		"action", action,
		"trace_id", traceID,
		"result", *result,
		"status", *statusCode,
		"duration_ms", durationMs,
	)
}
