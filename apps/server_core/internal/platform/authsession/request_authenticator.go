package authsession

import (
	"context"
	"errors"
	"net/http"
	"time"

	platformauth "metalshopping/server_core/internal/platform/auth"
)

type RequestAuthenticator struct {
	config Config
	store  interface {
		GetActiveSession(ctx context.Context, sessionID string, now time.Time) (Session, error)
	}
	now func() time.Time
}

func NewRequestAuthenticator(config Config, store interface {
	GetActiveSession(ctx context.Context, sessionID string, now time.Time) (Session, error)
}) *RequestAuthenticator {
	return &RequestAuthenticator{
		config: config,
		store:  store,
		now:    defaultNow,
	}
}

func (a *RequestAuthenticator) AuthenticateRequest(ctx context.Context, r *http.Request) (platformauth.Principal, error) {
	if a == nil || a.store == nil {
		return platformauth.Principal{}, platformauth.ErrUnauthenticated
	}

	cookie, err := r.Cookie(a.config.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return platformauth.Principal{}, platformauth.ErrUnauthenticated
		}
		return platformauth.Principal{}, err
	}

	session, err := a.store.GetActiveSession(ctx, cookie.Value, a.now())
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) || errors.Is(err, ErrSessionInvalidated) {
			return platformauth.Principal{}, platformauth.ErrUnauthenticated
		}
		return platformauth.Principal{}, err
	}

	return platformauth.Principal{
		SubjectID: session.SubjectID,
		TenantID:  session.TenantID,
		Email:     session.Email,
		Name:      session.DisplayName,
	}, nil
}
