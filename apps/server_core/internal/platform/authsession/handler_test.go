package authsession

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateCSRFCookieRequestAcceptsMatchingHeaderAndCookie(t *testing.T) {
	handler := NewHandler(nil, Config{
		CSRFCookieName: "ms_web_csrf",
		CSRFHeaderName: "X-CSRF-Token",
		CSRFHMACSecret: "test-secret",
		CSRFTTL:        30 * time.Minute,
	}, []string{"http://127.0.0.1:5173"})
	handler.csrf.now = func() time.Time { return time.Date(2026, 3, 18, 14, 0, 0, 0, time.UTC) }
	token, _, err := handler.csrf.Issue("session-123", "tenant-123")
	if err != nil {
		t.Fatalf("expected csrf token to be issued: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/refresh", nil)
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	request.Header.Set("X-CSRF-Token", token)
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: token})

	recorder := httptest.NewRecorder()
	if !handler.validateCSRFCookieRequest(recorder, request, "session-123", "tenant-123") {
		t.Fatalf("expected csrf validation to succeed, got status %d", recorder.Code)
	}
}

func TestValidateCSRFCookieRequestRejectsMissingHeader(t *testing.T) {
	handler := NewHandler(nil, Config{
		CSRFCookieName: "ms_web_csrf",
		CSRFHeaderName: "X-CSRF-Token",
		CSRFHMACSecret: "test-secret",
		CSRFTTL:        30 * time.Minute,
	}, []string{"http://127.0.0.1:5173"})
	handler.csrf.now = func() time.Time { return time.Date(2026, 3, 18, 14, 0, 0, 0, time.UTC) }
	token, _, err := handler.csrf.Issue("session-123", "tenant-123")
	if err != nil {
		t.Fatalf("expected csrf token to be issued: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/logout", nil)
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: token})

	recorder := httptest.NewRecorder()
	if handler.validateCSRFCookieRequest(recorder, request, "session-123", "tenant-123") {
		t.Fatal("expected csrf validation to fail without header")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestValidateCSRFCookieRequestRejectsUnexpectedOrigin(t *testing.T) {
	handler := NewHandler(nil, Config{
		CSRFCookieName: "ms_web_csrf",
		CSRFHeaderName: "X-CSRF-Token",
		CSRFHMACSecret: "test-secret",
		CSRFTTL:        30 * time.Minute,
	}, []string{"http://127.0.0.1:5173"})
	handler.csrf.now = func() time.Time { return time.Date(2026, 3, 18, 14, 0, 0, 0, time.UTC) }
	token, _, err := handler.csrf.Issue("session-123", "tenant-123")
	if err != nil {
		t.Fatalf("expected csrf token to be issued: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/logout", nil)
	request.Header.Set("Origin", "http://malicious.local")
	request.Header.Set("X-CSRF-Token", token)
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: token})

	recorder := httptest.NewRecorder()
	if handler.validateCSRFCookieRequest(recorder, request, "session-123", "tenant-123") {
		t.Fatal("expected csrf validation to fail for invalid origin")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestValidateCSRFCookieRequestRejectsTokenForDifferentSession(t *testing.T) {
	handler := NewHandler(nil, Config{
		CSRFCookieName: "ms_web_csrf",
		CSRFHeaderName: "X-CSRF-Token",
		CSRFHMACSecret: "test-secret",
		CSRFTTL:        30 * time.Minute,
	}, []string{"http://127.0.0.1:5173"})
	handler.csrf.now = func() time.Time { return time.Date(2026, 3, 18, 14, 0, 0, 0, time.UTC) }
	token, _, err := handler.csrf.Issue("session-123", "tenant-123")
	if err != nil {
		t.Fatalf("expected csrf token to be issued: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/logout", nil)
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	request.Header.Set("X-CSRF-Token", token)
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: token})

	recorder := httptest.NewRecorder()
	if handler.validateCSRFCookieRequest(recorder, request, "session-other", "tenant-123") {
		t.Fatal("expected csrf validation to fail for mismatched session")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestValidateCSRFCookieRequestRejectsExpiredToken(t *testing.T) {
	handler := NewHandler(nil, Config{
		CSRFCookieName: "ms_web_csrf",
		CSRFHeaderName: "X-CSRF-Token",
		CSRFHMACSecret: "test-secret",
		CSRFTTL:        30 * time.Minute,
	}, []string{"http://127.0.0.1:5173"})
	handler.csrf.now = func() time.Time { return time.Date(2026, 3, 18, 14, 0, 0, 0, time.UTC) }
	token, _, err := handler.csrf.Issue("session-123", "tenant-123")
	if err != nil {
		t.Fatalf("expected csrf token to be issued: %v", err)
	}
	handler.csrf.now = func() time.Time { return time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC) }

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/logout", nil)
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	request.Header.Set("X-CSRF-Token", token)
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: token})

	recorder := httptest.NewRecorder()
	if handler.validateCSRFCookieRequest(recorder, request, "session-123", "tenant-123") {
		t.Fatal("expected csrf validation to fail for expired token")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}
