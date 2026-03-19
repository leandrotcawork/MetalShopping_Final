package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/suppliers/application"
	"metalshopping/server_core/internal/modules/suppliers/ports"
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
	mux.HandleFunc("/api/v1/suppliers/directory", h.handleDirectory)
	mux.HandleFunc("/api/v1/suppliers/directory/", h.handleDirectoryByCode)
	mux.HandleFunc("/api/v1/suppliers/manifests", h.handleManifests)
}

type upsertDirectoryRequest struct {
	SupplierCode  string `json:"supplierCode"`
	SupplierLabel string `json:"supplierLabel"`
	ExecutionKind string `json:"executionKind"`
	LookupPolicy  string `json:"lookupPolicy"`
	Enabled       *bool  `json:"enabled"`
}

type setDirectoryEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type createManifestRequest struct {
	SupplierCode string          `json:"supplierCode"`
	Family       string          `json:"family"`
	Config       json.RawMessage `json:"config"`
}

func (h *Handler) handleDirectory(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		onlyEnabled := parseQueryBool(r, "enabled_only", false)
		rows, err := h.service.ListDirectory(r.Context(), tenantID, onlyEnabled)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list suppliers directory", requestTraceID(r))
			return
		}
		payload := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			payload = append(payload, mapDirectorySupplier(row))
		}
		writeJSON(w, http.StatusOK, map[string]any{"rows": payload})
	case http.MethodPut:
		var req upsertDirectoryRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "SUPPLIERS_DIRECTORY_INVALID", "Invalid suppliers directory payload", requestTraceID(r))
			return
		}
		enabled := true
		if req.Enabled != nil {
			enabled = *req.Enabled
		}
		item, err := h.service.UpsertDirectorySupplier(r.Context(), tenantID, ports.UpsertDirectorySupplierInput{
			SupplierCode:  req.SupplierCode,
			SupplierLabel: req.SupplierLabel,
			ExecutionKind: req.ExecutionKind,
			LookupPolicy:  req.LookupPolicy,
			Enabled:       enabled,
		})
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "SUPPLIERS_DIRECTORY_INVALID", err.Error(), requestTraceID(r))
			return
		}
		writeJSON(w, http.StatusOK, mapDirectorySupplier(item))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleDirectoryByCode(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	trimmedPath := strings.TrimPrefix(r.URL.Path, "/api/v1/suppliers/directory/")
	if !strings.HasSuffix(trimmedPath, "/enabled") {
		http.NotFound(w, r)
		return
	}
	supplierCode := strings.TrimSuffix(trimmedPath, "/enabled")
	supplierCode = strings.TrimSpace(strings.TrimSuffix(supplierCode, "/"))
	if supplierCode == "" || strings.Contains(supplierCode, "/") {
		http.NotFound(w, r)
		return
	}

	var req setDirectoryEnabledRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "SUPPLIERS_DIRECTORY_INVALID", "Invalid suppliers enablement payload", requestTraceID(r))
		return
	}

	item, err := h.service.SetDirectorySupplierEnabled(r.Context(), tenantID, supplierCode, req.Enabled)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			writeAPIError(w, http.StatusNotFound, "SUPPLIERS_DIRECTORY_NOT_FOUND", "Supplier not found", requestTraceID(r))
			return
		}
		writeAPIError(w, http.StatusBadRequest, "SUPPLIERS_DIRECTORY_INVALID", err.Error(), requestTraceID(r))
		return
	}
	writeJSON(w, http.StatusOK, mapDirectorySupplier(item))
}

func (h *Handler) handleManifests(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := authenticatedTenantID(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		limit := parseQueryInt64(r, "limit", 50)
		offset := parseQueryInt64(r, "offset", 0)
		supplierCode := strings.TrimSpace(r.URL.Query().Get("supplier_code"))
		list, err := h.service.ListDriverManifests(r.Context(), tenantID, supplierCode, limit, offset)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list supplier manifests", requestTraceID(r))
			return
		}
		rows := make([]map[string]any, 0, len(list.Rows))
		for _, item := range list.Rows {
			rows = append(rows, mapDriverManifest(item))
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"rows": rows,
			"paging": map[string]any{
				"offset":   list.Offset,
				"limit":    list.Limit,
				"returned": len(rows),
				"total":    list.Total,
			},
		})
	case http.MethodPost:
		principal, principalOK := platformauth.PrincipalFromContext(r.Context())
		if !principalOK {
			writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
			return
		}

		var req createManifestRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "SUPPLIERS_MANIFEST_INVALID", "Invalid supplier manifest payload", requestTraceID(r))
			return
		}

		createdBy := strings.TrimSpace(principal.Email)
		if createdBy == "" {
			createdBy = strings.TrimSpace(principal.SubjectID)
		}
		if createdBy == "" {
			createdBy = "unknown"
		}

		item, err := h.service.CreateDriverManifest(r.Context(), tenantID, ports.CreateDriverManifestInput{
			SupplierCode: req.SupplierCode,
			Family:       req.Family,
			ConfigJSON:   req.Config,
			CreatedBy:    createdBy,
		})
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "SUPPLIERS_MANIFEST_INVALID", err.Error(), requestTraceID(r))
			return
		}
		writeJSON(w, http.StatusCreated, mapDriverManifest(item))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func mapDirectorySupplier(item ports.DirectorySupplier) map[string]any {
	return map[string]any{
		"supplierCode":  item.SupplierCode,
		"supplierLabel": item.SupplierLabel,
		"executionKind": item.ExecutionKind,
		"lookupPolicy":  item.LookupPolicy,
		"enabled":       item.Enabled,
		"updatedAt":     item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapDriverManifest(item ports.DriverManifest) map[string]any {
	return map[string]any{
		"manifestId":       item.ManifestID,
		"supplierCode":     item.SupplierCode,
		"versionNumber":    item.VersionNumber,
		"family":           item.Family,
		"config":           json.RawMessage(item.ConfigJSON),
		"validationStatus": item.ValidationStatus,
		"validationErrors": json.RawMessage(item.ValidationErrors),
		"isActive":         item.IsActive,
		"createdBy":        item.CreatedBy,
		"createdAt":        item.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":        item.UpdatedAt.UTC().Format(time.RFC3339),
	}
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

func parseQueryBool(r *http.Request, key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(r.URL.Query().Get(key)))
	switch raw {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}

func authenticatedTenantID(w http.ResponseWriter, r *http.Request) (string, bool) {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return "", false
	}
	tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
		return "", false
	}
	return tenant.ID, true
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
