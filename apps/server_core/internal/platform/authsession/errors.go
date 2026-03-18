package authsession

import "errors"

var (
	ErrSessionNotFound       = errors.New("auth session not found")
	ErrSessionExpired        = errors.New("auth session expired")
	ErrSessionInvalidated    = errors.New("auth session invalidated")
	ErrLoginStateNotFound    = errors.New("auth login state not found")
	ErrLoginStateExpired     = errors.New("auth login state expired")
	ErrWebSessionDisabled    = errors.New("auth web session disabled")
	ErrMissingSessionCookie  = errors.New("auth missing session cookie")
	ErrMissingStateCookie    = errors.New("auth missing state cookie")
	ErrStateMismatch         = errors.New("auth login state mismatch")
	ErrOIDCConfigIncomplete  = errors.New("auth oidc configuration incomplete")
	ErrOIDCTokenExchange     = errors.New("auth oidc token exchange failed")
	ErrReturnTargetRejected  = errors.New("auth return target rejected")
	ErrGovernanceResolution  = errors.New("auth session governance resolution failed")
	ErrMissingSessionTimeout = errors.New("auth session timeout governance is missing")
)
