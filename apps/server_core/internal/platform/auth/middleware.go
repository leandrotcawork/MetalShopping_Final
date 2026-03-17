package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Authenticator interface {
	Authenticate(ctx context.Context, accessToken string) (Principal, error)
}

type Middleware struct {
	authenticator Authenticator
	publicPaths   map[string]struct{}
}

func NewMiddleware(authenticator Authenticator, publicPaths []string) *Middleware {
	pathSet := make(map[string]struct{}, len(publicPaths))
	for _, path := range publicPaths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		pathSet[trimmed] = struct{}{}
	}

	return &Middleware{
		authenticator: authenticator,
		publicPaths:   pathSet,
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if m == nil || m.authenticator == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		accessToken, err := ExtractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
			return
		}

		principal, err := m.authenticator.Authenticate(r.Context(), accessToken)
		if err != nil {
			if errors.Is(err, ErrUnauthenticated) {
				writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication failed", requestTraceID(r))
				return
			}
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authentication subsystem failure", requestTraceID(r))
			return
		}

		if err := principal.Validate(); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid authenticated principal", requestTraceID(r))
			return
		}

		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
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
