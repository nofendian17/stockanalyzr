package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"stockanalyzr/pkg/cache"
)

func TestNewRedis_Success(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer s.Close()

	dsn := "redis://" + s.Addr() + "/0"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	r, err := cache.NewRedis(ctx, dsn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r == nil {
		t.Fatal("expected Redis instance, got nil")
	}
	if r.Client == nil {
		t.Fatal("expected redis client, got nil")
	}

	err = r.Shutdown()
	if err != nil {
		t.Errorf("expected no error on shutdown, got %v", err)
	}
}

func TestNewRedis_InvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	dsn := "invalid-url"
	r, err := cache.NewRedis(ctx, dsn)
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
	if r != nil {
		t.Errorf("expected nil Redis instance, got %v", r)
	}
}

func TestNewRedis_PingFailure(t *testing.T) {
	// Provide a valid DSN but point it to a port where nothing is listening
	dsn := "redis://127.0.0.1:0/0"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	r, err := cache.NewRedis(ctx, dsn)
	if err == nil {
		t.Fatal("expected error due to ping failure, got nil")
	}
	if r != nil {
		t.Errorf("expected nil Redis instance, got %v", r)
	}
}

func TestHealthCheck_Success(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer s.Close()

	dsn := "redis://" + s.Addr() + "/0"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	r, err := cache.NewRedis(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create redis instance: %v", err)
	}
	defer r.Shutdown()

	err = r.HealthCheck()
	if err != nil {
		t.Errorf("expected no error on healthcheck, got %v", err)
	}
}

func TestHealthCheck_Failure(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	dsn := "redis://" + s.Addr() + "/0"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	r, err := cache.NewRedis(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create redis instance: %v", err)
	}

	// Close miniredis to simulate failure
	s.Close()

	// Need a small delay or retry for standard go-redis client to realize connection is broken?
	// or we can just do one call and verify it fails.
	time.Sleep(time.Millisecond * 100) // Wait briefly for OS to release socket

	err = r.HealthCheck()
	if err == nil {
		t.Fatal("expected error on healthcheck after server shutdown, got nil")
	}
}
