package unit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestJWTAuthenticatorAuthenticatesValidToken(t *testing.T) {
	t.Setenv("MS_AUTH_JWT_ALGORITHM", "HS256")
	t.Setenv("MS_AUTH_JWT_ISSUER", "metalshopping-test")
	t.Setenv("MS_AUTH_JWT_AUDIENCE", "metalshopping-api")
	t.Setenv("MS_AUTH_JWT_HMAC_SECRET", "super-secret")

	authenticator, err := auth.NewJWTAuthenticatorFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	token := buildHS256JWT(t, "super-secret", map[string]any{
		"sub":       "user-123",
		"iss":       "metalshopping-test",
		"aud":       "metalshopping-api",
		"tenant_id": "tenant-xyz",
		"email":     "user@example.com",
		"name":      "Example User",
		"exp":       time.Now().Add(5 * time.Minute).Unix(),
	})

	principal, err := authenticator.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("expected token to authenticate, got %v", err)
	}
	if principal.SubjectID != "user-123" || principal.TenantID != "tenant-xyz" {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestJWTAuthenticatorRejectsInvalidIssuer(t *testing.T) {
	t.Setenv("MS_AUTH_JWT_ALGORITHM", "HS256")
	t.Setenv("MS_AUTH_JWT_ISSUER", "metalshopping-prod")
	t.Setenv("MS_AUTH_JWT_AUDIENCE", "metalshopping-api")
	t.Setenv("MS_AUTH_JWT_HMAC_SECRET", "super-secret")

	authenticator, err := auth.NewJWTAuthenticatorFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	token := buildHS256JWT(t, "super-secret", map[string]any{
		"sub":       "user-123",
		"iss":       "wrong-issuer",
		"aud":       "metalshopping-api",
		"tenant_id": "tenant-xyz",
		"exp":       time.Now().Add(5 * time.Minute).Unix(),
	})

	if _, err := authenticator.Authenticate(context.Background(), token); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
}

func TestAuthenticatorFromEnvKeepsStaticDefault(t *testing.T) {
	t.Setenv("MS_AUTH_MODE", "")
	t.Setenv("MS_AUTH_STATIC_BEARER_TOKEN", "local-token")
	t.Setenv("MS_AUTH_STATIC_SUBJECT_ID", "bootstrap-user")

	authenticator, err := auth.NewAuthenticatorFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := authenticator.(*auth.StaticBearerAuthenticator); !ok {
		t.Fatalf("expected static authenticator by default, got %T", authenticator)
	}
}

func buildHS256JWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]any{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		t.Fatalf("marshal jwt header: %v", err)
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal jwt claims: %v", err)
	}

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	message := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return message + "." + signature
}
