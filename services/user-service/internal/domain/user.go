package domain

import "time"

// User is the core entity for user-service domain.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Disabled     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TokenPair holds access and refresh tokens with their expiry times.
type TokenPair struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}
