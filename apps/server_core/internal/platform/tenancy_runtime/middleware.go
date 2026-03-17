package tenancy_runtime

import (
	"encoding/json"
	"net/http"
	"strings"

	platformauth "metalshopping/server_core/internal/platform/auth"
)

type Middleware struct {
	publicPaths map[string]struct{}
}

func NewMiddleware(publicPaths []string) *Middleware {
	pathSet := make(map[string]struct{}, len(publicPaths))
	for _, path := range publicPaths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		pathSet[trimmed] = struct{}{}
	}

	return &Middleware{publicPaths: pathSet}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if m == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		principal, ok := platformauth.PrincipalFromContext(r.Context())
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
			return
		}

		tenantID := strings.TrimSpace(principal.TenantID)
		if tenantID == "" {
			writeAPIError(w, http.StatusForbidden, "TENANCY_FORBIDDEN", "Tenant context required", requestTraceID(r))
			return
		}

		next.ServeHTTP(w, r.WithContext(WithTenant(r.Context(), Tenant{ID: tenantID})))
	})
}

func (m *Middleware) isPublicPath(path string) bool {
	_, ok := m.publicPaths[path]
	return ok
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiErrorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: map[string]any{},
			TraceID: traceID,
		},
	})
}
