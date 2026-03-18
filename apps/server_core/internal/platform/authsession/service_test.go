package authsession

import (
	"context"
	"strings"
	"testing"
	"time"

	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	platformauth "metalshopping/server_core/internal/platform/auth"
	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
	"metalshopping/server_core/internal/platform/governance/threshold_resolver"
)

type fakeSessionStorage struct {
	loginStates      map[string]LoginState
	sessions         map[string]Session
	lastCreatedState LoginState
	lastCreated      Session
}

func (f *fakeSessionStorage) CreateLoginState(_ context.Context, loginState LoginState) error {
	f.lastCreatedState = loginState
	if f.loginStates == nil {
		f.loginStates = map[string]LoginState{}
	}
	f.loginStates[loginState.LoginStateID] = loginState
	return nil
}

func (f *fakeSessionStorage) GetLoginState(_ context.Context, loginStateID string) (LoginState, error) {
	state, ok := f.loginStates[loginStateID]
	if !ok {
		return LoginState{}, ErrLoginStateNotFound
	}
	return state, nil
}

func (f *fakeSessionStorage) DeleteLoginState(_ context.Context, loginStateID string) error {
	delete(f.loginStates, loginStateID)
	return nil
}

func (f *fakeSessionStorage) CreateSession(_ context.Context, session Session) error {
	f.lastCreated = session
	if f.sessions == nil {
		f.sessions = map[string]Session{}
	}
	f.sessions[session.SessionID] = session
	return nil
}

func (f *fakeSessionStorage) GetActiveSession(_ context.Context, sessionID string, now time.Time) (Session, error) {
	session, ok := f.sessions[sessionID]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	if session.InvalidatedAt != nil {
		return Session{}, ErrSessionInvalidated
	}
	if now.After(session.IdleTimeoutExpiresAt) || now.After(session.AbsoluteTimeoutExpiresAt) {
		return Session{}, ErrSessionExpired
	}
	return session, nil
}

func (f *fakeSessionStorage) RotateSession(_ context.Context, currentSessionID string, next Session, invalidatedAt time.Time) error {
	current := f.sessions[currentSessionID]
	current.InvalidatedAt = &invalidatedAt
	f.sessions[currentSessionID] = current
	f.sessions[next.SessionID] = next
	return nil
}

func (f *fakeSessionStorage) InvalidateSession(_ context.Context, sessionID string, _ string, invalidatedAt time.Time) error {
	session, ok := f.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}
	session.InvalidatedAt = &invalidatedAt
	f.sessions[sessionID] = session
	return nil
}

type fakeTokenExchanger struct {
	response TokenExchangeResponse
}

func (f fakeTokenExchanger) ExchangeCode(context.Context, string, string) (TokenExchangeResponse, error) {
	return f.response, nil
}

type fakeTokenValidator struct {
	principal platformauth.Principal
}

func (f fakeTokenValidator) Authenticate(context.Context, string) (platformauth.Principal, error) {
	return f.principal, nil
}

type fakeRoleReader struct {
	roles []iamdomain.Role
}

func (f fakeRoleReader) RolesByUserID(context.Context, string) ([]iamdomain.Role, error) {
	return f.roles, nil
}

func newPolicyResolver() *RuntimePolicyResolver {
	registry := governancebootstrap.NewRegistry()
	enabled := true
	return NewRuntimePolicyResolver(
		feature_flags.NewResolver(registry, map[string]feature_flags.ScopeValues{
			governancebootstrap.AuthWebSessionEnabledKey: {Global: &enabled},
		}),
		threshold_resolver.NewResolver(registry, map[string]threshold_resolver.ScopeValues{
			governancebootstrap.AuthSessionIdleTimeoutMinutesKey:     {Global: floatPtr(30)},
			governancebootstrap.AuthSessionAbsoluteTimeoutMinutesKey: {Global: floatPtr(480)},
		}),
		"local",
	)
}

func TestServiceStartLoginCreatesStateAndRedirect(t *testing.T) {
	store := &fakeSessionStorage{}
	config := Config{
		AuthorizationEndpoint: "https://idp.example.com/authorize",
		TokenEndpoint:         "https://idp.example.com/token",
		ClientID:              "metalshopping-web",
		RedirectURI:           "http://127.0.0.1:8080/api/v1/auth/session/callback",
		Scopes:                []string{"openid", "profile", "email"},
		DefaultReturnTo:       "/products",
		LoginStateTTL:         10 * time.Minute,
		HTTPTimeout:           10 * time.Second,
	}
	service := NewService(config, store, NewOIDCClient(config), fakeTokenValidator{}, newPolicyResolver(), fakeRoleReader{}, nil)
	service.now = func() time.Time { return time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC) }

	redirectURL, state, expiresAt, err := service.StartLogin(context.Background(), "/products")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state == "" {
		t.Fatal("expected generated state")
	}
	if store.lastCreatedState.LoginStateID == "" || store.lastCreatedState.CodeVerifier == "" {
		t.Fatalf("expected login state to be persisted, got %+v", store.lastCreatedState)
	}
	if !strings.Contains(redirectURL, "state="+state) {
		t.Fatalf("expected redirect url to contain state, got %s", redirectURL)
	}
	if expiresAt.Sub(service.now()) != 10*time.Minute {
		t.Fatalf("expected 10 minute expiry, got %v", expiresAt.Sub(service.now()))
	}
}

func TestServiceCompleteLoginCreatesSession(t *testing.T) {
	store := &fakeSessionStorage{
		loginStates: map[string]LoginState{
			"state-1": {
				LoginStateID: "state-1",
				CodeVerifier: "verifier-1",
				ReturnTo:     "/products",
				ExpiresAt:    time.Date(2026, 3, 18, 12, 10, 0, 0, time.UTC),
			},
		},
	}
	service := NewService(
		Config{DefaultReturnTo: "/products"},
		store,
		fakeTokenExchanger{response: TokenExchangeResponse{IDToken: "jwt-token"}},
		fakeTokenValidator{principal: platformauth.Principal{
			SubjectID: "user-1",
			TenantID:  "tenant-1",
			Email:     "user@example.com",
			Name:      "Example User",
		}},
		newPolicyResolver(),
		fakeRoleReader{roles: []iamdomain.Role{iamdomain.RolePricingManager}},
		nil,
	)
	service.now = func() time.Time { return time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC) }

	session, returnTo, err := service.CompleteLogin(context.Background(), "state-1", "state-1", "code-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session.SessionID == "" {
		t.Fatal("expected generated session id")
	}
	if returnTo != "/products" {
		t.Fatalf("expected return target /products, got %q", returnTo)
	}
	if _, ok := store.sessions[session.SessionID]; !ok {
		t.Fatalf("expected session to be stored, got %+v", store.sessions)
	}
}

func TestServiceGetSessionStateResolvesRolesAndCapabilities(t *testing.T) {
	store := &fakeSessionStorage{
		sessions: map[string]Session{
			"session-1": {
				SessionID:                "session-1",
				SubjectID:                "user-1",
				TenantID:                 "tenant-1",
				Email:                    "user@example.com",
				DisplayName:              "Example User",
				IssuedAt:                 time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC),
				LastSeenAt:               time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC),
				IdleTimeoutExpiresAt:     time.Date(2026, 3, 18, 12, 30, 0, 0, time.UTC),
				AbsoluteTimeoutExpiresAt: time.Date(2026, 3, 18, 20, 0, 0, 0, time.UTC),
			},
		},
	}
	service := NewService(
		Config{},
		store,
		nil,
		nil,
		newPolicyResolver(),
		fakeRoleReader{roles: []iamdomain.Role{iamdomain.RoleViewer}},
		fakeAuthorizer{},
	)
	service.now = func() time.Time { return time.Date(2026, 3, 18, 12, 1, 0, 0, time.UTC) }

	state, err := service.GetSessionState(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.UserID != "user-1" || state.TenantID != "tenant-1" {
		t.Fatalf("unexpected session state: %+v", state)
	}
	if len(state.Roles) != 1 || state.Roles[0] != string(iamdomain.RoleViewer) {
		t.Fatalf("unexpected roles: %+v", state.Roles)
	}
	if len(state.Capabilities) == 0 {
		t.Fatalf("expected capabilities to be resolved, got %+v", state.Capabilities)
	}
}

type fakeAuthorizer struct{}

func (fakeAuthorizer) Can(role iamdomain.Role, permission iamdomain.Permission) bool {
	return role == iamdomain.RoleViewer && permission == iamdomain.PermCatalogRead
}

func floatPtr(value float64) *float64 {
	return &value
}
