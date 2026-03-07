package rediscache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"stockanalyzr/services/user-service/internal/domain"
)

const (
	userKeyPrefix = "user:"
	userCacheTTL  = 15 * time.Minute
)

// UserCache is Redis implementation of domain.UserCache.
type UserCache struct {
	client *redis.Client
}

// Compile-time interface compliance check.
var _ domain.UserCache = (*UserCache)(nil)

func NewUserCache(client *redis.Client) *UserCache {
	return &UserCache{client: client}
}

func (c *UserCache) Get(ctx context.Context, userID string) (domain.User, error) {
	data, err := c.client.Get(ctx, userKey(userID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("cache get: %w", err)
	}

	var user domain.User
	if err := json.Unmarshal(data, &user); err != nil {
		return domain.User{}, fmt.Errorf("cache unmarshal: %w", err)
	}

	return user, nil
}

func (c *UserCache) Set(ctx context.Context, user domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	if err := c.client.Set(ctx, userKey(user.ID), data, userCacheTTL).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	return nil
}

func (c *UserCache) Delete(ctx context.Context, userID string) error {
	if err := c.client.Del(ctx, userKey(userID)).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}

	return nil
}

func userKey(userID string) string {
	return userKeyPrefix + userID
}
