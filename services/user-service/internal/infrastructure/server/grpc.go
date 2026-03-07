package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"stockanalyzr/pkg/logger"
)

// GRPCServer wraps a grpc.Server with lifecycle management.
type GRPCServer struct {
	server *grpc.Server
}

// NewGRPCServer creates a new gRPC server with the given options.
func NewGRPCServer(opts ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{
		server: grpc.NewServer(opts...),
	}
}

// Server returns the underlying grpc.Server for service registration.
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// ListenAndServe starts the gRPC server and blocks until a shutdown signal is received.
// It handles graceful shutdown via SIGINT/SIGTERM.
func (s *GRPCServer) ListenAndServe(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// Graceful shutdown goroutine
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		logger.Info("shutting down gRPC server...")
		s.server.GracefulStop()
	}()

	logger.Infof("gRPC server listening on :%s", port)
	return s.server.Serve(listener)
}

// Shutdown gracefully stops the gRPC server. Implements do.Shutdownable.
func (s *GRPCServer) Shutdown() error {
	s.server.GracefulStop()
	return nil
}
