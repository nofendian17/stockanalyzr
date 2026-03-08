package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"stockanalyzr/pkg/logger"
)

// HTTPServer wraps a Gin server with lifecycle management.
type HTTPServer struct {
	router *gin.Engine
}

// NewHTTPServer creates a new HTTP server with default middleware.
func NewHTTPServer() *HTTPServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Default middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(corsMiddleware())
	r.Use(requestIDMiddleware())

	// Custom error handler
	r.NoRoute(handleNoRoute)

	return &HTTPServer{router: r}
}

// Router returns the underlying Gin router for route registration.
func (s *HTTPServer) Router() *gin.Engine {
	return s.router
}

// ListenAndServe starts the HTTP server and blocks until a shutdown signal is received.
func (s *HTTPServer) ListenAndServe(port string) error {
	// Graceful shutdown goroutine
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	logger.Infof("HTTP server listening on :%s", port)
	return srv.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *HTTPServer) Shutdown() error {
	return nil // Placeholder, actual shutdown handled by http.Server in ListenAndServe
}

// corsMiddleware handles CORS headers.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// requestIDMiddleware generates a unique request ID for each request.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("X-Request-ID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID generates a simple request ID (could be improved with UUID).
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// handleNoRoute handles 404 errors.
func handleNoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":   true,
		"message": "not found",
		"code":    http.StatusNotFound,
	})
}
