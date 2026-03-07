package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Redis wraps a go-redis Client with lifecycle methods.
// Implements do.Shutdownable and do.Healthchecker for samber/do integration.
type Redis struct {
	Client *redis.Client
}

// NewRedis creates a new Redis client and verifies connectivity.
func NewRedis(ctx context.Context, dsn string) (*Redis, error) {
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("cache: invalid dsn: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("cache: failed to ping: %w", err)
	}

	return &Redis{Client: client}, nil
}

// Shutdown closes the Redis client. Implements do.Shutdownable.
func (r *Redis) Shutdown() error {
	return r.Client.Close()
}

// HealthCheck verifies Redis connectivity. Implements do.Healthchecker.
func (r *Redis) HealthCheck() error {
	return r.Client.Ping(context.Background()).Err()
}
