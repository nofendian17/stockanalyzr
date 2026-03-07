package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"stockanalyzr/services/user-service/internal/domain"
)

var publicMethods = map[string]struct{}{
	"/user.v1.UserService/Register": {},
	"/user.v1.UserService/Login":    {},
}

// AuthUnaryInterceptor validates bearer JWT for protected methods.
func AuthUnaryInterceptor(tokenManager domain.TokenManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := grpcHeaderFromIncomingContext(ctx, "authorization")
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		if !strings.HasPrefix(strings.ToLower(md), "bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		token := strings.TrimSpace(md[len("Bearer "):])
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		userID, err := tokenManager.ValidateAccessToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		authData := AuthData{
			UserID:      userID,
			AccessToken: token,
		}
		ctx = context.WithValue(ctx, authDataKey, authData)
		return handler(ctx, req)
	}
}

func grpcHeaderFromIncomingContext(ctx context.Context, key string) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	values := md.Get(key)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}
