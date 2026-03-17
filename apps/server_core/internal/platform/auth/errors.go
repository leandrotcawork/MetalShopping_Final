package auth

import "errors"

var (
	ErrMissingBearerToken   = errors.New("auth missing bearer token")
	ErrInvalidAuthorization = errors.New("auth invalid authorization header")
	ErrUnauthenticated      = errors.New("auth unauthenticated")
)
