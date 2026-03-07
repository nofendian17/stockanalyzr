package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres wraps a pgxpool.Pool with lifecycle methods.
// Implements do.Shutdownable and do.Healthchecker for samber/do integration.
type Postgres struct {
	Pool *pgxpool.Pool
}

// NewPostgres creates a new connection pool and verifies connectivity.
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("database: invalid dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("database: failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: failed to ping: %w", err)
	}

	return &Postgres{Pool: pool}, nil
}

// Shutdown closes the connection pool. Implements do.Shutdownable.
func (p *Postgres) Shutdown() error {
	p.Pool.Close()
	return nil
}

// HealthCheck verifies database connectivity. Implements do.Healthchecker.
func (p *Postgres) HealthCheck() error {
	return p.Pool.Ping(context.Background())
}
