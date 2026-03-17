package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

func TestTenancyMiddlewareSkipsPublicPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := tenancy_runtime.NewMiddleware([]string{"/health/live"})
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTenancyMiddlewareRequiresAuthenticatedPrincipal(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := tenancy_runtime.NewMiddleware(nil)
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/items", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestTenancyMiddlewareRequiresTenantID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := tenancy_runtime.NewMiddleware(nil)
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/items", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "user-1"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestTenancyMiddlewareInjectsTenantContext(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
		if !ok {
			t.Fatal("expected tenant in context")
		}
		if tenant.ID != "tenant-xyz" {
			t.Fatalf("expected tenant-xyz, got %q", tenant.ID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := tenancy_runtime.NewMiddleware(nil)
	h := middleware.Wrap(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/items", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{
		SubjectID: "user-1",
		TenantID:  "tenant-xyz",
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
