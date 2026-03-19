package authsession

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AuthorizationEndpoint    string
	TokenEndpoint            string
	ClientID                 string
	ClientSecret             string
	RedirectURI              string
	Scopes                   []string
	ForceLoginPrompt         bool
	DefaultReturnTo          string
	PostLoginRedirectBaseURL string
	SessionCookieName        string
	StateCookieName          string
	CSRFCookieName           string
	CSRFHeaderName           string
	CSRFHMACSecret           string
	CookieDomain             string
	CookiePath               string
	CookieSecure             bool
	CookieSameSite           http.SameSite
	LoginStateTTL            time.Duration
	CSRFTTL                  time.Duration
	HTTPTimeout              time.Duration
}

func ConfigFromEnv(environment string) (Config, error) {
	config := Config{
		AuthorizationEndpoint:    strings.TrimSpace(os.Getenv("MS_AUTH_OIDC_AUTHORIZATION_URL")),
		TokenEndpoint:            strings.TrimSpace(os.Getenv("MS_AUTH_OIDC_TOKEN_URL")),
		ClientID:                 strings.TrimSpace(os.Getenv("MS_AUTH_OIDC_CLIENT_ID")),
		ClientSecret:             strings.TrimSpace(os.Getenv("MS_AUTH_OIDC_CLIENT_SECRET")),
		RedirectURI:              strings.TrimSpace(os.Getenv("MS_AUTH_OIDC_REDIRECT_URI")),
		PostLoginRedirectBaseURL: strings.TrimSpace(os.Getenv("MS_AUTH_WEB_POST_LOGIN_BASE_URL")),
		ForceLoginPrompt:         parseBoolEnv("MS_AUTH_OIDC_FORCE_LOGIN_PROMPT", false),
		CookieDomain:             strings.TrimSpace(os.Getenv("MS_AUTH_WEB_COOKIE_DOMAIN")),
		CookiePath:               firstNonEmpty(strings.TrimSpace(os.Getenv("MS_AUTH_WEB_COOKIE_PATH")), "/"),
		SessionCookieName:        firstNonEmpty(strings.TrimSpace(os.Getenv("MS_AUTH_WEB_SESSION_COOKIE_NAME")), "ms_web_session"),
		StateCookieName:          firstNonEmpty(strings.TrimSpace(os.Getenv("MS_AUTH_WEB_STATE_COOKIE_NAME")), "ms_web_login_state"),
		CSRFCookieName:           firstNonEmpty(strings.TrimSpace(os.Getenv("MS_AUTH_WEB_CSRF_COOKIE_NAME")), "ms_web_csrf"),
		CSRFHeaderName:           firstNonEmpty(strings.TrimSpace(os.Getenv("MS_AUTH_WEB_CSRF_HEADER_NAME")), "X-CSRF-Token"),
		CSRFHMACSecret:           strings.TrimSpace(os.Getenv("MS_AUTH_WEB_CSRF_HMAC_SECRET")),
		DefaultReturnTo:          firstNonEmpty(strings.TrimSpace(os.Getenv("MS_AUTH_WEB_DEFAULT_RETURN_TO")), "/products"),
		CookieSecure:             strings.ToLower(strings.TrimSpace(environment)) != "local",
		CookieSameSite:           parseSameSite(os.Getenv("MS_AUTH_WEB_COOKIE_SAMESITE")),
		LoginStateTTL:            parseDurationMinutesEnv("MS_AUTH_WEB_LOGIN_STATE_TTL_MINUTES", 10),
		CSRFTTL:                  parseDurationMinutesEnv("MS_AUTH_WEB_CSRF_TTL_MINUTES", 30),
		HTTPTimeout:              parseDurationSecondsEnv("MS_AUTH_OIDC_HTTP_TIMEOUT_SECONDS", 10),
	}
	if config.PostLoginRedirectBaseURL == "" && strings.EqualFold(strings.TrimSpace(environment), "local") {
		config.PostLoginRedirectBaseURL = "http://127.0.0.1:5173"
	}

	rawScopes := strings.TrimSpace(os.Getenv("MS_AUTH_OIDC_SCOPES"))
	if rawScopes == "" {
		config.Scopes = []string{"openid", "profile", "email"}
	} else {
		config.Scopes = strings.Fields(rawScopes)
	}

	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.AuthorizationEndpoint) == "" {
		return fmt.Errorf("%w: authorization endpoint is required", ErrOIDCConfigIncomplete)
	}
	if strings.TrimSpace(c.TokenEndpoint) == "" {
		return fmt.Errorf("%w: token endpoint is required", ErrOIDCConfigIncomplete)
	}
	if strings.TrimSpace(c.ClientID) == "" {
		return fmt.Errorf("%w: client id is required", ErrOIDCConfigIncomplete)
	}
	if strings.TrimSpace(c.RedirectURI) == "" {
		return fmt.Errorf("%w: redirect uri is required", ErrOIDCConfigIncomplete)
	}
	if strings.TrimSpace(c.PostLoginRedirectBaseURL) != "" {
		parsed, err := url.Parse(c.PostLoginRedirectBaseURL)
		if err != nil || !parsed.IsAbs() || parsed.Host == "" {
			return fmt.Errorf("%w: post login redirect base url must be an absolute URL", ErrOIDCConfigIncomplete)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("%w: post login redirect base url must use http or https", ErrOIDCConfigIncomplete)
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return fmt.Errorf("%w: post login redirect base url must not include query or fragment", ErrOIDCConfigIncomplete)
		}
	}
	if strings.TrimSpace(c.CSRFCookieName) == "" {
		return fmt.Errorf("%w: csrf cookie name is required", ErrOIDCConfigIncomplete)
	}
	if strings.TrimSpace(c.CSRFHeaderName) == "" {
		return fmt.Errorf("%w: csrf header name is required", ErrOIDCConfigIncomplete)
	}
	if strings.TrimSpace(c.CSRFHMACSecret) == "" {
		return fmt.Errorf("%w: csrf hmac secret is required", ErrOIDCConfigIncomplete)
	}
	if len(c.Scopes) == 0 {
		return fmt.Errorf("%w: scopes are required", ErrOIDCConfigIncomplete)
	}
	if c.LoginStateTTL <= 0 {
		return fmt.Errorf("%w: login state ttl must be positive", ErrOIDCConfigIncomplete)
	}
	if c.CSRFTTL <= 0 {
		return fmt.Errorf("%w: csrf ttl must be positive", ErrOIDCConfigIncomplete)
	}
	if c.HTTPTimeout <= 0 {
		return fmt.Errorf("%w: oidc http timeout must be positive", ErrOIDCConfigIncomplete)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseSameSite(raw string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func parseDurationMinutesEnv(key string, fallbackMinutes int) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return time.Duration(fallbackMinutes) * time.Minute
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return time.Duration(fallbackMinutes) * time.Minute
	}
	return time.Duration(parsed) * time.Minute
}

func parseDurationSecondsEnv(key string, fallbackSeconds int) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return time.Duration(fallbackSeconds) * time.Second
	}
	return time.Duration(parsed) * time.Second
}

func parseBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
