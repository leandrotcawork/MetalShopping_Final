package domain

import "errors"

var (
	ErrUserIDRequired = errors.New("iam user id is required")
	ErrInvalidRole    = errors.New("iam role is invalid")
	ErrActorRequired  = errors.New("iam actor is required")
	ErrUserNotFound   = errors.New("iam user not found")
)
