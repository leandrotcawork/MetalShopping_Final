package authsession

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
}

type TokenExchanger interface {
	ExchangeCode(ctx context.Context, code string, codeVerifier string) (TokenExchangeResponse, error)
}

type OIDCClient struct {
	config     Config
	httpClient *http.Client
}

func NewOIDCClient(config Config) *OIDCClient {
	return &OIDCClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}
}

func (c *OIDCClient) AuthorizationURL(state string, codeVerifier string) string {
	codeChallengeBytes := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeBytes[:])
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", c.config.ClientID)
	values.Set("redirect_uri", c.config.RedirectURI)
	values.Set("scope", strings.Join(c.config.Scopes, " "))
	values.Set("state", state)
	values.Set("code_challenge", codeChallenge)
	values.Set("code_challenge_method", "S256")
	return c.config.AuthorizationEndpoint + "?" + values.Encode()
}

func (c *OIDCClient) ExchangeCode(ctx context.Context, code string, codeVerifier string) (TokenExchangeResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", c.config.ClientID)
	form.Set("redirect_uri", c.config.RedirectURI)
	form.Set("code_verifier", codeVerifier)
	if strings.TrimSpace(c.config.ClientSecret) != "" {
		form.Set("client_secret", c.config.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenExchangeResponse{}, fmt.Errorf("%w: %v", ErrOIDCTokenExchange, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return TokenExchangeResponse{}, fmt.Errorf("%w: %v", ErrOIDCTokenExchange, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return TokenExchangeResponse{}, fmt.Errorf("%w: status %d: %s", ErrOIDCTokenExchange, response.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload TokenExchangeResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return TokenExchangeResponse{}, fmt.Errorf("%w: decode token response: %v", ErrOIDCTokenExchange, err)
	}
	return payload, nil
}

func sanitizeReturnTarget(raw string, fallback string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}
	if !strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "//") {
		return "", ErrReturnTargetRejected
	}
	return trimmed, nil
}

func newExpiresAt(now time.Time, ttl time.Duration) time.Time {
	return now.UTC().Add(ttl)
}

func truncateTime(now time.Time) time.Time {
	return now.UTC().Round(0)
}

func defaultNow() time.Time {
	return time.Now().UTC()
}
