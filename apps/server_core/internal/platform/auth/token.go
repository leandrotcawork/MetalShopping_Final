package auth

import (
	"strings"
)

func ExtractBearerToken(headerValue string) (string, error) {
	raw := strings.TrimSpace(headerValue)
	if raw == "" {
		return "", ErrMissingBearerToken
	}

	parts := strings.SplitN(raw, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidAuthorization
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", ErrInvalidAuthorization
	}

	return token, nil
}
