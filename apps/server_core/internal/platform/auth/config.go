package auth

import (
	"os"
	"strings"
)

const (
	ModeStatic = "static"
	ModeJWT    = "jwt"
)

func NewAuthenticatorFromEnv() (Authenticator, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MS_AUTH_MODE"))) {
	case "", ModeStatic:
		return NewStaticBearerAuthenticatorFromEnv()
	case ModeJWT:
		return NewJWTAuthenticatorFromEnv()
	default:
		return nil, ErrInvalidAuthenticationMode
	}
}
