package authsession

import (
	"net/url"
	"testing"
)

func TestAuthorizationURLAddsPromptWhenForceLoginPromptEnabled(t *testing.T) {
	client := NewOIDCClient(Config{
		AuthorizationEndpoint: "http://127.0.0.1:18081/realms/metalshopping/protocol/openid-connect/auth",
		ClientID:              "metalshopping-web",
		RedirectURI:           "http://127.0.0.1:8080/api/v1/auth/session/callback",
		Scopes:                []string{"openid", "profile", "email"},
		ForceLoginPrompt:      true,
	})

	redirectURL := client.AuthorizationURL("state-123", "code-verifier-xyz")
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		t.Fatalf("expected authorization url to be valid: %v", err)
	}
	if got := parsed.Query().Get("prompt"); got != "login" {
		t.Fatalf("expected prompt=login, got %q", got)
	}
}

func TestAuthorizationURLOmitsPromptWhenForceLoginPromptDisabled(t *testing.T) {
	client := NewOIDCClient(Config{
		AuthorizationEndpoint: "http://127.0.0.1:18081/realms/metalshopping/protocol/openid-connect/auth",
		ClientID:              "metalshopping-web",
		RedirectURI:           "http://127.0.0.1:8080/api/v1/auth/session/callback",
		Scopes:                []string{"openid", "profile", "email"},
		ForceLoginPrompt:      false,
	})

	redirectURL := client.AuthorizationURL("state-123", "code-verifier-xyz")
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		t.Fatalf("expected authorization url to be valid: %v", err)
	}
	if got := parsed.Query().Get("prompt"); got != "" {
		t.Fatalf("expected prompt to be omitted, got %q", got)
	}
}
