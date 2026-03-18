package authsession

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateCSRFCookieRequestAcceptsMatchingHeaderAndCookie(t *testing.T) {
	handler := NewHandler(nil, Config{
		CSRFCookieName: "ms_web_csrf",
		CSRFHeaderName: "X-CSRF-Token",
	}, []string{"http://127.0.0.1:5173"})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/refresh", nil)
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	request.Header.Set("X-CSRF-Token", "csrf-123")
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: "csrf-123"})

	recorder := httptest.NewRecorder()
	if !handler.validateCSRFCookieRequest(recorder, request) {
		t.Fatalf("expected csrf validation to succeed, got status %d", recorder.Code)
	}
}

func TestValidateCSRFCookieRequestRejectsMissingHeader(t *testing.T) {
	handler := NewHandler(nil, Config{
		CSRFCookieName: "ms_web_csrf",
		CSRFHeaderName: "X-CSRF-Token",
	}, []string{"http://127.0.0.1:5173"})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/logout", nil)
	request.Header.Set("Origin", "http://127.0.0.1:5173")
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: "csrf-123"})

	recorder := httptest.NewRecorder()
	if handler.validateCSRFCookieRequest(recorder, request) {
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
	}, []string{"http://127.0.0.1:5173"})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session/logout", nil)
	request.Header.Set("Origin", "http://malicious.local")
	request.Header.Set("X-CSRF-Token", "csrf-123")
	request.AddCookie(&http.Cookie{Name: "ms_web_csrf", Value: "csrf-123"})

	recorder := httptest.NewRecorder()
	if handler.validateCSRFCookieRequest(recorder, request) {
		t.Fatal("expected csrf validation to fail for invalid origin")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}
