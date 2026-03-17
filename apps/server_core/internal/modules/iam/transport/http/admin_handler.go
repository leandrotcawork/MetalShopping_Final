package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"metalshopping/server_core/internal/modules/iam/application"
	"metalshopping/server_core/internal/modules/iam/domain"
	platformauth "metalshopping/server_core/internal/platform/auth"
)

type AdminHandler struct {
	service       *application.AdminService
	authorization *application.AuthorizationService
}

type UpsertRoleAssignmentRequest struct {
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func NewAdminHandler(service *application.AdminService, authorization *application.AuthorizationService) *AdminHandler {
	return &AdminHandler{service: service, authorization: authorization}
}

func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/iam/users/", h.handleUpsertRoleAssignment)
}

func (h *AdminHandler) handleUpsertRoleAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/iam/users/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || parts[1] != "roles" {
		writeAPIError(w, http.StatusNotFound, "ROUTE_NOT_FOUND", "Route not found", requestTraceID(r))
		return
	}

	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return
	}
	allowed, err := h.authorization.HasPermission(r.Context(), principal.SubjectID, domain.PermIAMManageRoles)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", requestTraceID(r))
		return
	}
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", requestTraceID(r))
		return
	}

	var req UpsertRoleAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", requestTraceID(r))
		return
	}

	err = h.service.UpsertRoleAssignment(r.Context(), application.UpsertRoleAssignmentCommand{
		UserID:      strings.TrimSpace(parts[0]),
		DisplayName: req.DisplayName,
		Role:        req.Role,
		AssignedBy:  principal.SubjectID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRole), errors.Is(err, domain.ErrUserIDRequired), errors.Is(err, domain.ErrActorRequired):
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upsert role assignment", requestTraceID(r))
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":      strings.TrimSpace(parts[0]),
		"display_name": strings.TrimSpace(req.DisplayName),
		"role":         strings.ToLower(strings.TrimSpace(req.Role)),
	})
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
