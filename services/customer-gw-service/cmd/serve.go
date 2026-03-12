package cmd

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"stockanalyzr/pkg/logger"
	_ "stockanalyzr/services/customer-gw-service/docs"
	"stockanalyzr/services/customer-gw-service/internal/config"
	"stockanalyzr/services/customer-gw-service/internal/delivery/http"
	"stockanalyzr/services/customer-gw-service/internal/domain"
	"stockanalyzr/services/customer-gw-service/internal/infrastructure/server"
	"stockanalyzr/services/customer-gw-service/internal/middleware"
	"stockanalyzr/services/customer-gw-service/internal/provider"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API Gateway server",
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
	handler := do.MustInvoke[*http.UserHandler](injector)
	uc := do.MustInvoke[domain.UserUsecase](injector)

	// Build HTTP server
	httpSrv := server.NewHTTPServer()
	r := httpSrv.Router()

	// Setup routes
	setupRoutes(r, handler, uc, cfg.JWTSecret)

	logger.Infof("customer-gw-service starting on :%s", cfg.HTTPPort)

	// Graceful shutdown on exit
	defer func() {
		if err := injector.Shutdown(); err != nil {
			logger.Errorf("shutdown error: %v", err)
		}
	}()

	_ = context.Background()
	return httpSrv.ListenAndServe(cfg.HTTPPort)
}

// setupRoutes configures all HTTP routes.
func setupRoutes(r *gin.Engine, handler *http.UserHandler, uc domain.UserUsecase, jwtSecret string) {
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")

	// Public routes
	v1.POST("/auth/register", handler.Register)
	v1.POST("/auth/login", handler.Login)
	v1.POST("/auth/refresh", handler.RefreshToken)

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.JWTMiddleware(jwtSecret, uc))

	// User routes - specific routes must be registered before parameterized routes
	protected.GET("/users/me", handler.GetProfile)
	protected.PUT("/users/me", handler.UpdateProfile)
	protected.DELETE("/users/me", handler.DeactivateAccount)

	// Auth routes
	protected.POST("/auth/logout", handler.Logout)
}
