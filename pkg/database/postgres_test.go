package database_test

import (
	"context"
	"testing"
	"time"

	"stockanalyzr/pkg/database"
)

func TestNewPostgres_InvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	dsn := "invalid-dsn" // Should trigger pgxpool.ParseConfig error
	p, err := database.NewPostgres(ctx, dsn)
	if err == nil {
		t.Fatal("expected error parsing invalid DSN, got nil")
	}
	if p != nil {
		t.Errorf("expected Postgres instance to be nil, got %v", p)
	}
}

func TestNewPostgres_PingFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	// Valid DSN format but points to nowhere to force ping failure
	dsn := "postgres://user:password@127.0.0.1:0/dbname"
	p, err := database.NewPostgres(ctx, dsn)
	if err == nil {
		t.Fatal("expected error due to ping failure, got nil")
	}
	if p != nil {
		t.Errorf("expected Postgres instance to be nil, got %v", p)
	}
}
