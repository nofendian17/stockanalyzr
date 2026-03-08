package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"stockanalyzr/services/customer-gw-service/internal/domain"
)

const tokenBlacklistPrefix = "jwt_bl:"

// TokenRepository is Redis implementation of domain.TokenRepository.
type TokenRepository struct {
	client *redis.Client
}

// Compile-time interface compliance check.
var _ domain.TokenRepository = (*TokenRepository)(nil)

// NewTokenRepository creates a new TokenRepository.
func NewTokenRepository(client *redis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

// Blacklist adds a token to the blacklist with the specified TTL.
func (r *TokenRepository) Blacklist(ctx context.Context, token string, ttlSeconds int) error {
	key := tokenBlacklistPrefix + token
	if err := r.client.Set(ctx, key, "blacklisted", time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// IsBlacklisted checks if a token is in the blacklist.
func (r *TokenRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := tokenBlacklistPrefix + token
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	return exists > 0, nil
}
