package cmd

import (
	"context"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	userv1 "stockanalyzr/pkg/gen"
	"stockanalyzr/pkg/logger"
	"stockanalyzr/services/user-service/internal/config"
	transportgrpc "stockanalyzr/services/user-service/internal/delivery/grpc"
	"stockanalyzr/services/user-service/internal/domain"
	"stockanalyzr/services/user-service/internal/infrastructure/server"
	"stockanalyzr/services/user-service/internal/middleware"
	"stockanalyzr/services/user-service/internal/provider"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the gRPC server",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(_ *cobra.Command, _ []string) error {
	injector := do.New()

	// Register config eagerly
	cfg := config.Load()
	do.ProvideValue(injector, cfg)

	// Register all providers
	provider.Package(injector)

	// Resolve dependencies
	handler := do.MustInvoke[*transportgrpc.UserHandler](injector)
	tokenManager := do.MustInvoke[domain.TokenManager](injector)

	// Build gRPC server
	grpcSrv := server.NewGRPCServer(
		grpc.UnaryInterceptor(middleware.AuthUnaryInterceptor(tokenManager)),
	)
	userv1.RegisterUserServiceServer(grpcSrv.Server(), handler)

	logger.Infof("user-service starting on :%s", cfg.GRPCPort)

	// Graceful shutdown on exit
	defer func() {
		if err := injector.Shutdown(); err != nil {
			logger.Errorf("shutdown error: %v", err)
		}
	}()

	_ = context.Background()
	return grpcSrv.ListenAndServe(cfg.GRPCPort)
}
