package authsession

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	iamports "metalshopping/server_core/internal/modules/iam/ports"
	platformauth "metalshopping/server_core/internal/platform/auth"
)

type SessionState struct {
	UserID                   string
	TenantID                 string
	DisplayName              string
	Email                    string
	Roles                    []string
	Capabilities             []string
	IssuedAt                 time.Time
	ExpiresAt                time.Time
	IdleTimeoutExpiresAt     time.Time
	AbsoluteTimeoutExpiresAt time.Time
	SessionID                string
}

type Service struct {
	config         Config
	store          SessionStorage
	tokenExchanger TokenExchanger
	tokenValidator platformauth.Authenticator
	policyResolver *RuntimePolicyResolver
	roleReader     iamports.RoleAssignmentReader
	authorizer     iamdomain.Authorizer
	now            func() time.Time
}

func NewService(
	config Config,
	store SessionStorage,
	tokenExchanger TokenExchanger,
	tokenValidator platformauth.Authenticator,
	policyResolver *RuntimePolicyResolver,
	roleReader iamports.RoleAssignmentReader,
	authorizer iamdomain.Authorizer,
) *Service {
	return &Service{
		config:         config,
		store:          store,
		tokenExchanger: tokenExchanger,
		tokenValidator: tokenValidator,
		policyResolver: policyResolver,
		roleReader:     roleReader,
		authorizer:     authorizer,
		now:            defaultNow,
	}
}

func (s *Service) StartLogin(ctx context.Context, requestedReturnTo string) (redirectURL string, state string, expiresAt time.Time, err error) {
	if s == nil || s.store == nil || s.tokenExchanger == nil {
		return "", "", time.Time{}, ErrOIDCConfigIncomplete
	}

	state, err = randomToken(32)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate auth login state: %w", err)
	}
	codeVerifier, err := randomToken(48)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate auth code verifier: %w", err)
	}
	returnTo, err := sanitizeReturnTarget(requestedReturnTo, s.config.DefaultReturnTo)
	if err != nil {
		return "", "", time.Time{}, err
	}

	expiresAt = newExpiresAt(s.now(), s.config.LoginStateTTL)
	if err := s.store.CreateLoginState(ctx, LoginState{
		LoginStateID: state,
		CodeVerifier: codeVerifier,
		ReturnTo:     returnTo,
		ExpiresAt:    expiresAt,
	}); err != nil {
		return "", "", time.Time{}, err
	}

	oidcClient, ok := s.tokenExchanger.(*OIDCClient)
	if !ok {
		return "", "", time.Time{}, ErrOIDCConfigIncomplete
	}
	return oidcClient.AuthorizationURL(state, codeVerifier), state, expiresAt, nil
}

func (s *Service) CompleteLogin(ctx context.Context, state string, stateCookie string, code string) (Session, string, error) {
	if strings.TrimSpace(state) == "" || strings.TrimSpace(code) == "" {
		return Session{}, "", ErrLoginStateNotFound
	}
	if strings.TrimSpace(stateCookie) == "" {
		return Session{}, "", ErrMissingStateCookie
	}
	if strings.TrimSpace(state) != strings.TrimSpace(stateCookie) {
		return Session{}, "", ErrStateMismatch
	}

	loginState, err := s.store.GetLoginState(ctx, state)
	if err != nil {
		return Session{}, "", err
	}
	if s.now().UTC().After(loginState.ExpiresAt.UTC()) {
		_ = s.store.DeleteLoginState(ctx, state)
		return Session{}, "", ErrLoginStateExpired
	}

	tokenResponse, err := s.tokenExchanger.ExchangeCode(ctx, code, loginState.CodeVerifier)
	if err != nil {
		return Session{}, "", err
	}
	tokenToValidate := strings.TrimSpace(tokenResponse.IDToken)
	if tokenToValidate == "" {
		tokenToValidate = strings.TrimSpace(tokenResponse.AccessToken)
	}
	principal, err := s.tokenValidator.Authenticate(ctx, tokenToValidate)
	if err != nil {
		return Session{}, "", err
	}

	policy, err := s.policyResolver.Resolve(ctx, principal.TenantID)
	if err != nil {
		return Session{}, "", err
	}
	if !policy.Enabled {
		return Session{}, "", ErrWebSessionDisabled
	}

	now := truncateTime(s.now())
	sessionID, err := randomToken(32)
	if err != nil {
		return Session{}, "", fmt.Errorf("generate auth session id: %w", err)
	}
	session := Session{
		SessionID:                sessionID,
		SubjectID:                principal.SubjectID,
		TenantID:                 principal.TenantID,
		Email:                    principal.Email,
		DisplayName:              principal.Name,
		IssuedAt:                 now,
		LastSeenAt:               now,
		IdleTimeoutExpiresAt:     now.Add(policy.IdleTimeout),
		AbsoluteTimeoutExpiresAt: now.Add(policy.AbsoluteTimeout),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return Session{}, "", err
	}
	if err := s.store.DeleteLoginState(ctx, state); err != nil {
		return Session{}, "", err
	}
	return session, loginState.ReturnTo, nil
}

func (s *Service) GetSessionState(ctx context.Context, sessionID string) (SessionState, error) {
	session, err := s.store.GetActiveSession(ctx, sessionID, s.now())
	if err != nil {
		return SessionState{}, err
	}

	roles, capabilities, err := s.resolveRolesAndCapabilities(ctx, session.SubjectID)
	if err != nil {
		return SessionState{}, err
	}
	expiresAt := session.IdleTimeoutExpiresAt
	if session.AbsoluteTimeoutExpiresAt.Before(expiresAt) {
		expiresAt = session.AbsoluteTimeoutExpiresAt
	}
	return SessionState{
		UserID:                   session.SubjectID,
		TenantID:                 session.TenantID,
		DisplayName:              session.DisplayName,
		Email:                    session.Email,
		Roles:                    roles,
		Capabilities:             capabilities,
		IssuedAt:                 session.IssuedAt,
		ExpiresAt:                expiresAt,
		IdleTimeoutExpiresAt:     session.IdleTimeoutExpiresAt,
		AbsoluteTimeoutExpiresAt: session.AbsoluteTimeoutExpiresAt,
		SessionID:                session.SessionID,
	}, nil
}

func (s *Service) RefreshSession(ctx context.Context, sessionID string) (SessionState, Session, error) {
	currentSession, err := s.store.GetActiveSession(ctx, sessionID, s.now())
	if err != nil {
		return SessionState{}, Session{}, err
	}
	policy, err := s.policyResolver.Resolve(ctx, currentSession.TenantID)
	if err != nil {
		return SessionState{}, Session{}, err
	}
	if !policy.Enabled {
		return SessionState{}, Session{}, ErrWebSessionDisabled
	}

	now := truncateTime(s.now())
	nextSessionID, err := randomToken(32)
	if err != nil {
		return SessionState{}, Session{}, fmt.Errorf("generate rotated auth session id: %w", err)
	}
	nextSession := Session{
		SessionID:                nextSessionID,
		SubjectID:                currentSession.SubjectID,
		TenantID:                 currentSession.TenantID,
		Email:                    currentSession.Email,
		DisplayName:              currentSession.DisplayName,
		IssuedAt:                 currentSession.IssuedAt,
		LastSeenAt:               now,
		IdleTimeoutExpiresAt:     now.Add(policy.IdleTimeout),
		AbsoluteTimeoutExpiresAt: currentSession.AbsoluteTimeoutExpiresAt,
	}
	if nextSession.AbsoluteTimeoutExpiresAt.Before(nextSession.IdleTimeoutExpiresAt) {
		nextSession.IdleTimeoutExpiresAt = nextSession.AbsoluteTimeoutExpiresAt
	}
	if err := s.store.RotateSession(ctx, currentSession.SessionID, nextSession, now); err != nil {
		return SessionState{}, Session{}, err
	}

	state, err := s.GetSessionState(ctx, nextSession.SessionID)
	if err != nil {
		return SessionState{}, Session{}, err
	}
	return state, nextSession, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	err := s.store.InvalidateSession(ctx, sessionID, "logout", s.now())
	if errors.Is(err, ErrSessionNotFound) {
		return err
	}
	return err
}

func (s *Service) resolveRolesAndCapabilities(ctx context.Context, subjectID string) ([]string, []string, error) {
	if s.roleReader == nil || s.authorizer == nil {
		return []string{}, []string{}, nil
	}
	roles, err := s.roleReader.RolesByUserID(ctx, subjectID)
	if err != nil {
		if errors.Is(err, iamdomain.ErrUserNotFound) {
			return []string{}, []string{}, nil
		}
		return nil, nil, fmt.Errorf("load iam roles for auth session: %w", err)
	}

	roleNames := make([]string, 0, len(roles))
	capabilitySet := make(map[string]struct{})
	for _, role := range roles {
		roleNames = append(roleNames, string(role))
		for _, permission := range iamdomain.AllPermissions() {
			if s.authorizer.Can(role, permission) {
				capabilitySet[string(permission)] = struct{}{}
			}
		}
	}
	sort.Strings(roleNames)

	capabilities := make([]string, 0, len(capabilitySet))
	for permission := range capabilitySet {
		capabilities = append(capabilities, permission)
	}
	sort.Strings(capabilities)
	return roleNames, capabilities, nil
}

func randomToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
