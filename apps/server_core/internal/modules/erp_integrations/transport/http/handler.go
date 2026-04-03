package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/application"
	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

// Handler exposes the erp_integrations application service over HTTP.
type Handler struct {
	svc *application.Service
}

// NewHandler creates a new Handler wrapping the given service.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all ERP integration routes on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/erp/instances", h.listInstances)
	mux.HandleFunc("POST /api/v1/erp/instances", h.createInstance)
	mux.HandleFunc("GET /api/v1/erp/instances/{instance_id}", h.getInstance)
	mux.HandleFunc("POST /api/v1/erp/instances/{instance_id}/runs", h.triggerRun)
	mux.HandleFunc("GET /api/v1/erp/instances/{instance_id}/runs", h.listRuns)
	mux.HandleFunc("GET /api/v1/erp/instances/{instance_id}/runs/{run_id}", h.getRun)
	mux.HandleFunc("GET /api/v1/erp/review-items", h.listReviewItems)
	mux.HandleFunc("POST /api/v1/erp/review-items/{review_id}/resolve", h.resolveReviewItem)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (h *Handler) listInstances(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}
	_ = principal

	limit, offset := parsePagination(r)

	items, err := h.svc.ListInstances(r.Context(), tenant.ID, limit, offset)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list ERP integration instances", requestTraceID(r))
		return
	}

	payload := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload = append(payload, mapInstance(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": payload})
}

type createInstanceRequest struct {
	ConnectorType   string   `json:"connector_type"`
	DisplayName     string   `json:"display_name"`
	ConnectionRef   string   `json:"connection_ref"`
	EnabledEntities []string `json:"enabled_entities"`
	SyncSchedule    *string  `json:"sync_schedule,omitempty"`
}

func (h *Handler) createInstance(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	var req createInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", requestTraceID(r))
		return
	}

	enabledEntities := make([]domain.EntityType, len(req.EnabledEntities))
	for i, e := range req.EnabledEntities {
		enabledEntities[i] = domain.EntityType(e)
	}

	instance, err := h.svc.CreateInstance(r.Context(), application.CreateInstanceCommand{
		TenantID:        tenant.ID,
		PrincipalID:     principal.SubjectID,
		ConnectorType:   domain.ConnectorType(req.ConnectorType),
		DisplayName:     req.DisplayName,
		ConnectionRef:   req.ConnectionRef,
		EnabledEntities: enabledEntities,
		SyncSchedule:    req.SyncSchedule,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapInstance(instance))
}

func (h *Handler) getInstance(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	instanceID := r.PathValue("instance_id")
	if strings.TrimSpace(instanceID) == "" {
		http.NotFound(w, r)
		return
	}

	instance, err := h.svc.GetInstance(r.Context(), tenant.ID, instanceID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, mapInstance(instance))
}

type triggerRunRequest struct {
	RunMode     string   `json:"run_mode"`
	EntityScope []string `json:"entity_scope"`
}

func (h *Handler) triggerRun(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	instanceID := r.PathValue("instance_id")
	if strings.TrimSpace(instanceID) == "" {
		http.NotFound(w, r)
		return
	}

	var req triggerRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", requestTraceID(r))
		return
	}

	entityScope := make([]domain.EntityType, len(req.EntityScope))
	for i, e := range req.EntityScope {
		entityScope[i] = domain.EntityType(e)
	}

	run, err := h.svc.TriggerRun(r.Context(), application.TriggerRunCommand{
		TenantID:    tenant.ID,
		PrincipalID: principal.SubjectID,
		InstanceID:  instanceID,
		RunMode:     domain.RunMode(req.RunMode),
		EntityScope: entityScope,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusAccepted, mapRun(run))
}

func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	instanceID := r.PathValue("instance_id")
	if strings.TrimSpace(instanceID) == "" {
		http.NotFound(w, r)
		return
	}

	limit, offset := parsePagination(r)

	items, err := h.svc.ListRuns(r.Context(), tenant.ID, instanceID, limit, offset)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list ERP sync runs", requestTraceID(r))
		return
	}

	payload := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload = append(payload, mapRun(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": payload})
}

func (h *Handler) getRun(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	runID := r.PathValue("run_id")
	if strings.TrimSpace(runID) == "" {
		http.NotFound(w, r)
		return
	}

	run, err := h.svc.GetRun(r.Context(), tenant.ID, runID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, mapRun(run))
}

func (h *Handler) listReviewItems(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	limit, offset := parsePagination(r)

	items, err := h.svc.ListReviewItems(r.Context(), tenant.ID, limit, offset)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list ERP review items", requestTraceID(r))
		return
	}

	payload := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload = append(payload, mapReviewItem(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": payload})
}

type resolveReviewRequest struct {
	Resolution string `json:"resolution"` // "resolved" or "dismissed"
	Note       string `json:"note,omitempty"`
}

func (h *Handler) resolveReviewItem(w http.ResponseWriter, r *http.Request) {
	principal, tenant, ok := h.requirePrincipalAndTenant(w, r)
	if !ok {
		return
	}

	reviewID := r.PathValue("review_id")
	if strings.TrimSpace(reviewID) == "" {
		http.NotFound(w, r)
		return
	}

	var req resolveReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", requestTraceID(r))
		return
	}

	item, err := h.svc.ResolveReview(r.Context(), application.ResolveReviewCommand{
		TenantID:    tenant.ID,
		PrincipalID: principal.SubjectID,
		ReviewID:    reviewID,
		Resolution:  domain.ReviewItemStatus(req.Resolution),
		Note:        req.Note,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, mapReviewItem(item))
}

// ---------------------------------------------------------------------------
// Auth / tenant helpers
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Error mapping
// ---------------------------------------------------------------------------

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	traceID := requestTraceID(r)
	switch {
	case errors.Is(err, domain.ErrInstanceNotFound),
		errors.Is(err, domain.ErrRunNotFound),
		errors.Is(err, domain.ErrReviewItemNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), traceID)
	case errors.Is(err, domain.ErrActiveInstanceExists),
		errors.Is(err, domain.ErrReviewAlreadyResolved):
		writeAPIError(w, http.StatusConflict, "CONFLICT", err.Error(), traceID)
	case errors.Is(err, domain.ErrIntegrationDisabled):
		writeAPIError(w, http.StatusForbidden, "INTEGRATION_DISABLED", err.Error(), traceID)
	case errors.Is(err, domain.ErrInvalidConnectorType),
		errors.Is(err, domain.ErrInvalidEntityType),
		errors.Is(err, domain.ErrInvalidRunMode),
		errors.Is(err, domain.ErrEmptyDisplayName),
		errors.Is(err, domain.ErrEmptyEnabledEntities),
		errors.Is(err, domain.ErrEmptyEntityScope),
		errors.Is(err, domain.ErrEmptyTenantID),
		errors.Is(err, domain.ErrInvalidInstanceStatus):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), traceID)
	default:
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", traceID)
	}
}

// ---------------------------------------------------------------------------
// Response mappers
// ---------------------------------------------------------------------------

func mapInstance(i *domain.IntegrationInstance) map[string]any {
	entities := make([]string, len(i.EnabledEntities))
	for j, e := range i.EnabledEntities {
		entities[j] = string(e)
	}
	m := map[string]any{
		"instance_id":      i.InstanceID,
		"tenant_id":        i.TenantID,
		"connector_type":   string(i.ConnectorType),
		"display_name":     i.DisplayName,
		"connection_ref":   i.ConnectionRef,
		"enabled_entities": entities,
		"status":           string(i.Status),
		"created_at":       i.CreatedAt.Format(time.RFC3339),
		"updated_at":       i.UpdatedAt.Format(time.RFC3339),
	}
	if i.SyncSchedule != nil {
		m["sync_schedule"] = *i.SyncSchedule
	}
	return m
}

func mapRun(r *domain.SyncRun) map[string]any {
	scope := make([]string, len(r.EntityScope))
	for i, e := range r.EntityScope {
		scope[i] = string(e)
	}
	m := map[string]any{
		"run_id":         r.RunID,
		"tenant_id":      r.TenantID,
		"instance_id":    r.InstanceID,
		"connector_type": string(r.ConnectorType),
		"run_mode":       string(r.RunMode),
		"entity_scope":   scope,
		"status":         string(r.Status),
		"promoted_count": r.PromotedCount,
		"warning_count":  r.WarningCount,
		"rejected_count": r.RejectedCount,
		"review_count":   r.ReviewCount,
		"created_at":     r.CreatedAt.Format(time.RFC3339),
	}
	if r.StartedAt != nil {
		m["started_at"] = r.StartedAt.Format(time.RFC3339)
	}
	if r.CompletedAt != nil {
		m["completed_at"] = r.CompletedAt.Format(time.RFC3339)
	}
	if r.FailureSummary != nil {
		m["failure_summary"] = *r.FailureSummary
	}
	return m
}

func mapReviewItem(item *domain.ReviewItem) map[string]any {
	m := map[string]any{
		"review_id":             item.ReviewID,
		"tenant_id":             item.TenantID,
		"instance_id":           item.InstanceID,
		"connector_type":        string(item.ConnectorType),
		"entity_type":           string(item.EntityType),
		"source_id":             item.SourceID,
		"run_id":                item.RunID,
		"severity":              string(item.Severity),
		"reason_code":           item.ReasonCode,
		"problem_summary":       item.ProblemSummary,
		"raw_payload_ref":       item.RawPayloadRef,
		"staging_snapshot":      decodeOptionalJSON(item.StagingSnapshot),
		"reconciliation_output": decodeOptionalJSON(item.ReconciliationOutput),
		"recommended_action":    item.RecommendedAction,
		"item_status":           string(item.ItemStatus),
		"created_at":            item.CreatedAt.Format(time.RFC3339),
	}
	if item.ResolvedAt != nil {
		m["resolved_at"] = item.ResolvedAt.Format(time.RFC3339)
	}
	if item.ResolvedBy != nil {
		m["resolved_by"] = *item.ResolvedBy
	}
	return m
}

func decodeOptionalJSON(raw *string) any {
	if raw == nil {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(*raw), &decoded); err != nil {
		return *raw
	}
	return decoded
}

// ---------------------------------------------------------------------------
// JSON / HTTP utilities
// ---------------------------------------------------------------------------

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

// parsePagination extracts limit and offset from the request query string.
// Default limit is 50, maximum is 200. Default offset is 0.
func parsePagination(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 200 {
		limit = 200
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}
