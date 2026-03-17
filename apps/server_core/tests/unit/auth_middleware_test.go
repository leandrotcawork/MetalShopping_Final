package unit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"metalshopping/server_core/internal/platform/auth"
)

type fakeAuthenticator struct {
	principal auth.Principal
	err       error
}

func (f fakeAuthenticator) Authenticate(context.Context, string) (auth.Principal, error) {
	if f.err != nil {
		return auth.Principal{}, f.err
	}
	return f.principal, nil
}

func TestAuthMiddlewareSkipsConfiguredPublicPaths(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.NewMiddleware(fakeAuthenticator{}, []string{"/health/live"})
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddlewareRejectsMissingBearerToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.NewMiddleware(fakeAuthenticator{}, []string{"/health/live"})
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/items", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewareRejectsUnauthenticatedToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.NewMiddleware(fakeAuthenticator{err: auth.ErrUnauthenticated}, nil)
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/items", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewarePassesPrincipalThroughContext(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("expected principal in request context")
		}
		if principal.SubjectID != "user-123" {
			t.Fatalf("expected subject user-123, got %q", principal.SubjectID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.NewMiddleware(fakeAuthenticator{
		principal: auth.Principal{
			SubjectID: "user-123",
			TenantID:  "tenant-xyz",
			Email:     "user@example.com",
			Name:      "Example User",
		},
	}, nil)
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/items", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddlewareReturnsInternalErrorForAuthenticatorFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.NewMiddleware(fakeAuthenticator{err: errors.New("jwks unavailable")}, nil)
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/items", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
