package domain

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredential  = errors.New("invalid credential")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrUserAlreadyDeleted = errors.New("user already deleted")
	ErrUserNotDeleted     = errors.New("user not deleted")
)
