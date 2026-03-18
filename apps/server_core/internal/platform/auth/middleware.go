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

type RequestAuthenticator interface {
	AuthenticateRequest(ctx context.Context, r *http.Request) (Principal, error)
}

type Middleware struct {
	requestAuthenticator RequestAuthenticator
	publicPaths          map[string]struct{}
}

func NewMiddleware(authenticator Authenticator, publicPaths []string) *Middleware {
	return NewMiddlewareWithRequestAuthenticator(NewBearerRequestAuthenticator(authenticator), publicPaths)
}

func NewMiddlewareWithRequestAuthenticator(requestAuthenticator RequestAuthenticator, publicPaths []string) *Middleware {
	pathSet := make(map[string]struct{}, len(publicPaths))
	for _, path := range publicPaths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		pathSet[trimmed] = struct{}{}
	}

	return &Middleware{
		requestAuthenticator: requestAuthenticator,
		publicPaths:          pathSet,
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if m == nil || m.requestAuthenticator == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		principal, err := m.requestAuthenticator.AuthenticateRequest(r.Context(), r)
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

type bearerRequestAuthenticator struct {
	authenticator Authenticator
}

func NewBearerRequestAuthenticator(authenticator Authenticator) RequestAuthenticator {
	if authenticator == nil {
		return nil
	}
	return &bearerRequestAuthenticator{authenticator: authenticator}
}

func (a *bearerRequestAuthenticator) AuthenticateRequest(ctx context.Context, r *http.Request) (Principal, error) {
	accessToken, err := ExtractBearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}
	return a.authenticator.Authenticate(ctx, accessToken)
}

type bearerOrCookieRequestAuthenticator struct {
	bearer RequestAuthenticator
	cookie RequestAuthenticator
}

func NewBearerOrCookieRequestAuthenticator(bearer RequestAuthenticator, cookie RequestAuthenticator) RequestAuthenticator {
	return &bearerOrCookieRequestAuthenticator{
		bearer: bearer,
		cookie: cookie,
	}
}

func (a *bearerOrCookieRequestAuthenticator) AuthenticateRequest(ctx context.Context, r *http.Request) (Principal, error) {
	if strings.TrimSpace(r.Header.Get("Authorization")) != "" {
		if a.bearer == nil {
			return Principal{}, ErrUnauthenticated
		}
		return a.bearer.AuthenticateRequest(ctx, r)
	}
	if a.cookie == nil {
		return Principal{}, ErrUnauthenticated
	}
	return a.cookie.AuthenticateRequest(ctx, r)
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
