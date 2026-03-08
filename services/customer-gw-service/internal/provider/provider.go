package provider

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"stockanalyzr/pkg/cache"
	userpb "stockanalyzr/pkg/gen"
	"stockanalyzr/services/customer-gw-service/internal/config"
	"stockanalyzr/services/customer-gw-service/internal/delivery/http"
	"stockanalyzr/services/customer-gw-service/internal/domain"
	"stockanalyzr/services/customer-gw-service/internal/infrastructure/persistence"
	"stockanalyzr/services/customer-gw-service/internal/usecase"
)

// Package registers all gateway service dependencies into the DI container.
var Package = do.Package(
	// --- Infrastructure: Cache (Redis) ---
	do.Lazy(func(i do.Injector) (*cache.Redis, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return cache.NewRedis(context.Background(), cfg.RedisDSN)
	}),

	// --- Token Repository ---
	do.Lazy(func(i do.Injector) (*persistence.TokenRepository, error) {
		rds := do.MustInvoke[*cache.Redis](i)
		return persistence.NewTokenRepository(rds.Client), nil
	}),
	do.Bind[*persistence.TokenRepository, domain.TokenRepository](),

	// --- gRPC Client: User Service ---
	do.Lazy(func(i do.Injector) (*grpc.ClientConn, error) {
		cfg := do.MustInvoke[*config.Config](i)
		conn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to user service: %w", err)
		}
		return conn, nil
	}),

	// --- User Service Client ---
	do.Lazy(func(i do.Injector) (userpb.UserServiceClient, error) {
		conn := do.MustInvoke[*grpc.ClientConn](i)
		return userpb.NewUserServiceClient(conn), nil
	}),

	// --- Usecase ---
	do.Lazy(func(i do.Injector) (*usecase.UserInteractor, error) {
		client := do.MustInvoke[userpb.UserServiceClient](i)
		tokenRepo := do.MustInvoke[domain.TokenRepository](i)
		cfg := do.MustInvoke[*config.Config](i)
		return usecase.NewUserUsecase(client, tokenRepo, cfg.JWTSecret), nil
	}),
	do.Bind[*usecase.UserInteractor, domain.UserUsecase](),

	// --- HTTP Handler ---
	do.Lazy(func(i do.Injector) (*http.UserHandler, error) {
		uc := do.MustInvoke[domain.UserUsecase](i)
		return http.NewUserHandler(uc), nil
	}),
)
