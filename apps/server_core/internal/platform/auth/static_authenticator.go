package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type StaticBearerAuthenticator struct {
	accessToken string
	principal   Principal
}

func NewStaticBearerAuthenticator(accessToken string, principal Principal) (*StaticBearerAuthenticator, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("auth static bearer token is required")
	}
	if err := principal.Validate(); err != nil {
		return nil, err
	}

	return &StaticBearerAuthenticator{
		accessToken: accessToken,
		principal:   principal,
	}, nil
}

func NewStaticBearerAuthenticatorFromEnv() (*StaticBearerAuthenticator, error) {
	return NewStaticBearerAuthenticator(
		os.Getenv("MS_AUTH_STATIC_BEARER_TOKEN"),
		Principal{
			SubjectID: strings.TrimSpace(os.Getenv("MS_AUTH_STATIC_SUBJECT_ID")),
			TenantID:  strings.TrimSpace(os.Getenv("MS_AUTH_STATIC_TENANT_ID")),
			Email:     strings.TrimSpace(os.Getenv("MS_AUTH_STATIC_EMAIL")),
			Name:      strings.TrimSpace(os.Getenv("MS_AUTH_STATIC_NAME")),
		},
	)
}

func (a *StaticBearerAuthenticator) Authenticate(_ context.Context, accessToken string) (Principal, error) {
	if strings.TrimSpace(accessToken) != a.accessToken {
		return Principal{}, ErrUnauthenticated
	}
	return a.principal, nil
}
