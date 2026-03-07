package domain

import "context"

//go:generate mockgen -source=security.go -destination=../mocks/mock_security.go -package=mocks

// PasswordHasher abstracts password hashing implementation.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

// TokenManager abstracts token creation and validation.
type TokenManager interface {
	CreateTokenPair(ctx context.Context, userID string) (TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (string, error)
	RevokeToken(ctx context.Context, token string) error
}
