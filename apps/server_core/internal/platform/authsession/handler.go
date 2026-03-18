package authsession

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	platformauth "metalshopping/server_core/internal/platform/auth"
)

type Handler struct {
	service *Service
	config  Config
	allowedOrigins map[string]struct{}
}

func NewHandler(service *Service, config Config, allowedOrigins []string) *Handler {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		originSet[trimmed] = struct{}{}
	}

	return &Handler{service: service, config: config, allowedOrigins: originSet}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/session/login", h.handleLogin)
	mux.HandleFunc("/api/v1/auth/session/callback", h.handleCallback)
	mux.HandleFunc("/api/v1/auth/session/me", h.handleMe)
	mux.HandleFunc("/api/v1/auth/session/refresh", h.handleRefresh)
	mux.HandleFunc("/api/v1/auth/session/logout", h.handleLogout)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	redirectURL, state, expiresAt, err := h.service.StartLogin(r.Context(), r.URL.Query().Get("return_to"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	SetLoginStateCookie(w, h.config, state, expiresAt)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	stateCookie, err := r.Cookie(h.config.StateCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		h.writeError(w, r, err)
		return
	}
	stateCookieValue := ""
	if stateCookie != nil {
		stateCookieValue = stateCookie.Value
	}

	session, returnTo, err := h.service.CompleteLogin(
		r.Context(),
		r.URL.Query().Get("state"),
		stateCookieValue,
		r.URL.Query().Get("code"),
	)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	SetSessionCookie(w, h.config, session.SessionID, session.IdleTimeoutExpiresAt)
	SetCSRFCookie(w, h.config, randomCSRFToken(), session.IdleTimeoutExpiresAt)
	ClearLoginStateCookie(w, h.config)
	http.Redirect(w, r, returnTo, http.StatusFound)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sessionID, ok := h.requireSessionCookie(w, r)
	if !ok {
		return
	}
	state, err := h.service.GetSessionState(r.Context(), sessionID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	SetCSRFCookie(w, h.config, randomCSRFToken(), state.IdleTimeoutExpiresAt)
	writeJSON(w, http.StatusOK, mapSessionState(state))
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !h.validateCSRFCookieRequest(w, r) {
		return
	}

	sessionID, ok := h.requireSessionCookie(w, r)
	if !ok {
		return
	}
	state, nextSession, err := h.service.RefreshSession(r.Context(), sessionID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	SetSessionCookie(w, h.config, nextSession.SessionID, nextSession.IdleTimeoutExpiresAt)
	SetCSRFCookie(w, h.config, randomCSRFToken(), nextSession.IdleTimeoutExpiresAt)
	writeJSON(w, http.StatusOK, mapSessionState(state))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !h.validateCSRFCookieRequest(w, r) {
		return
	}

	sessionID, ok := h.requireSessionCookie(w, r)
	if !ok {
		return
	}
	if err := h.service.Logout(r.Context(), sessionID); err != nil {
		h.writeError(w, r, err)
		return
	}

	ClearSessionCookie(w, h.config)
	ClearCSRFCookie(w, h.config)
	writeJSON(w, http.StatusOK, map[string]any{
		"logged_out": true,
	})
}

func (h *Handler) validateCSRFCookieRequest(w http.ResponseWriter, r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" && len(h.allowedOrigins) > 0 {
		if _, ok := h.allowedOrigins[origin]; !ok {
			writeAPIError(w, http.StatusForbidden, "CSRF_REJECTED", "Request origin is not allowed", requestTraceID(r))
			return false
		}
	}

	csrfCookie, err := r.Cookie(h.config.CSRFCookieName)
	if err != nil || strings.TrimSpace(csrfCookie.Value) == "" {
		writeAPIError(w, http.StatusForbidden, "CSRF_REJECTED", "Missing CSRF cookie", requestTraceID(r))
		return false
	}

	headerValue := strings.TrimSpace(r.Header.Get(h.config.CSRFHeaderName))
	if headerValue == "" {
		writeAPIError(w, http.StatusForbidden, "CSRF_REJECTED", "Missing CSRF header", requestTraceID(r))
		return false
	}
	if headerValue != strings.TrimSpace(csrfCookie.Value) {
		writeAPIError(w, http.StatusForbidden, "CSRF_REJECTED", "Invalid CSRF token", requestTraceID(r))
		return false
	}
	return true
}

func (h *Handler) requireSessionCookie(w http.ResponseWriter, r *http.Request) (string, bool) {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return "", false
	}

	cookie, err := r.Cookie(h.config.SessionCookieName)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return "", false
	}
	return strings.TrimSpace(cookie.Value), true
}

func mapSessionState(state SessionState) map[string]any {
	return map[string]any{
		"user_id":                     state.UserID,
		"tenant_id":                   state.TenantID,
		"display_name":                state.DisplayName,
		"email":                       state.Email,
		"roles":                       state.Roles,
		"capabilities":                state.Capabilities,
		"issued_at":                   state.IssuedAt.UTC(),
		"expires_at":                  state.ExpiresAt.UTC(),
		"idle_timeout_expires_at":     state.IdleTimeoutExpiresAt.UTC(),
		"absolute_timeout_expires_at": state.AbsoluteTimeoutExpiresAt.UTC(),
		"session_id":                  state.SessionID,
	}
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrMissingStateCookie), errors.Is(err, ErrStateMismatch), errors.Is(err, ErrLoginStateNotFound), errors.Is(err, ErrLoginStateExpired), errors.Is(err, ErrReturnTargetRejected):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestTraceID(r))
	case errors.Is(err, ErrWebSessionDisabled), errors.Is(err, ErrMissingSessionTimeout):
		writeAPIError(w, http.StatusForbidden, "GOVERNANCE_DISABLED", err.Error(), requestTraceID(r))
	case errors.Is(err, platformauth.ErrUnauthenticated), errors.Is(err, ErrMissingSessionCookie), errors.Is(err, ErrSessionNotFound), errors.Is(err, ErrSessionExpired), errors.Is(err, ErrSessionInvalidated):
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication failed", requestTraceID(r))
	case errors.Is(err, ErrOIDCTokenExchange):
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication failed", requestTraceID(r))
	default:
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Auth session flow failed", requestTraceID(r))
	}
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

func randomCSRFToken() string {
	token, err := randomToken(24)
	if err != nil {
		return "csrf-token-unavailable"
	}
	return token
}
