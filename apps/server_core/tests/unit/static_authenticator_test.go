package unit

import (
	"context"
	"errors"
	"testing"

	"metalshopping/server_core/internal/platform/auth"
)

func TestStaticBearerAuthenticatorAuthenticatesConfiguredToken(t *testing.T) {
	authenticator, err := auth.NewStaticBearerAuthenticator("token-123", auth.Principal{
		SubjectID: "admin-local",
		Name:      "Local Admin",
	})
	if err != nil {
		t.Fatalf("expected authenticator, got %v", err)
	}

	principal, err := authenticator.Authenticate(context.Background(), "token-123")
	if err != nil {
		t.Fatalf("expected authentication success, got %v", err)
	}
	if principal.SubjectID != "admin-local" {
		t.Fatalf("expected admin-local, got %q", principal.SubjectID)
	}
}

func TestStaticBearerAuthenticatorRejectsWrongToken(t *testing.T) {
	authenticator, err := auth.NewStaticBearerAuthenticator("token-123", auth.Principal{
		SubjectID: "admin-local",
		Name:      "Local Admin",
	})
	if err != nil {
		t.Fatalf("expected authenticator, got %v", err)
	}

	_, err = authenticator.Authenticate(context.Background(), "bad-token")
	if !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
}
