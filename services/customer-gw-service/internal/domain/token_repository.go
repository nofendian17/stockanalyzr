package domain

import "context"

// TokenRepository defines operations for token blacklist management.
type TokenRepository interface {
	// Blacklist adds a token to the blacklist with a TTL based on token expiration
	Blacklist(ctx context.Context, token string, ttlSeconds int) error
	// IsBlacklisted checks if a token is in the blacklist
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}
