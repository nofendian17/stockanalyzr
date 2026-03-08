package provider

import (
	"context"

	"github.com/samber/do/v2"

	"stockanalyzr/pkg/cache"
	"stockanalyzr/pkg/database"
	"stockanalyzr/pkg/validator"
	"stockanalyzr/services/user-service/internal/config"
	transportgrpc "stockanalyzr/services/user-service/internal/delivery/grpc"
	"stockanalyzr/services/user-service/internal/domain"
	rediscache "stockanalyzr/services/user-service/internal/infrastructure/cache"
	"stockanalyzr/services/user-service/internal/infrastructure/persistence"
	"stockanalyzr/services/user-service/internal/security"
	"stockanalyzr/services/user-service/internal/usecase"
)

// Package registers all service dependencies into the DI container.
var Package = do.Package(
	// --- Infrastructure: Database ---
	do.Lazy(func(i do.Injector) (*database.Postgres, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return database.NewPostgres(context.Background(), cfg.PostgresDSN)
	}),

	// --- Infrastructure: Cache ---
	do.Lazy(func(i do.Injector) (*cache.Redis, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return cache.NewRedis(context.Background(), cfg.RedisDSN)
	}),

	// --- Repository ---
	do.Lazy(func(i do.Injector) (*persistence.UserRepository, error) {
		pg := do.MustInvoke[*database.Postgres](i)
		return persistence.NewUserRepository(pg.Pool), nil
	}),
	do.Bind[*persistence.UserRepository, domain.UserRepository](),

	// --- Cache ---
	do.Lazy(func(i do.Injector) (*rediscache.UserCache, error) {
		rds := do.MustInvoke[*cache.Redis](i)
		return rediscache.NewUserCache(rds.Client), nil
	}),
	do.Bind[*rediscache.UserCache, domain.UserCache](),

	// --- Security ---
	do.Lazy(func(i do.Injector) (*security.BcryptHasher, error) {
		return security.NewBcryptHasher(), nil
	}),
	do.Bind[*security.BcryptHasher, domain.PasswordHasher](),

	do.Lazy(func(i do.Injector) (*security.JWTManager, error) {
		cfg := do.MustInvoke[*config.Config](i)
		rds := do.MustInvoke[*cache.Redis](i)
		return security.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAccessExpiryMinutes, cfg.JWTRefreshExpiryMinutes, rds.Client), nil
	}),
	do.Bind[*security.JWTManager, domain.TokenManager](),

	// --- Shared Utils ---
	do.Lazy(func(i do.Injector) (*validator.Validator, error) {
		return validator.New(), nil
	}),

	// --- Usecase ---
	do.Lazy(func(i do.Injector) (*usecase.UserInteractor, error) {
		return usecase.NewUserUsecase(
			do.MustInvoke[domain.UserRepository](i),
			do.MustInvoke[domain.UserCache](i),
			do.MustInvoke[domain.PasswordHasher](i),
			do.MustInvoke[domain.TokenManager](i),
			do.MustInvoke[*validator.Validator](i),
		), nil
	}),
	do.Bind[*usecase.UserInteractor, domain.UserUsecase](),

	// --- Handler ---
	do.Lazy(func(i do.Injector) (*transportgrpc.UserHandler, error) {
		return transportgrpc.NewUserHandler(do.MustInvoke[domain.UserUsecase](i)), nil
	}),
)
