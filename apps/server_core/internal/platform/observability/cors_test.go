package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	middleware := NewCORSMiddleware([]string{"http://127.0.0.1:5173"})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/portfolio", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:5173" {
		t.Fatalf("expected allow-origin header, got %q", got)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	middleware := NewCORSMiddleware([]string{"http://127.0.0.1:5173"})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("preflight should not reach next handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/products/portfolio", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}
