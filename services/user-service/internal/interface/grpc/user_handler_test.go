package grpc_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "stockanalyzr/pkg/gen"
	"stockanalyzr/services/user-service/internal/domain"
	transportgrpc "stockanalyzr/services/user-service/internal/interface/grpc"
	"stockanalyzr/services/user-service/internal/mocks"
)

func newTestHandler(t *testing.T) (*transportgrpc.UserHandler, *mocks.MockUserUsecase) {
	ctrl := gomock.NewController(t)
	uc := mocks.NewMockUserUsecase(ctrl)
	handler := transportgrpc.NewUserHandler(uc)
	return handler, uc
}

func TestRegisterHandler_Success(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().Register(ctx, "test@example.com", "password123", "Test User").Return(domain.User{
		ID:       "user-123",
		Email:    "test@example.com",
		FullName: "Test User",
	}, nil)

	resp, err := handler.Register(ctx, &userv1.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		FullName: "Test User",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.User.Id != "user-123" {
		t.Errorf("expected user-123, got %s", resp.User.Id)
	}
}

func TestRegisterHandler_InvalidInput(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().Register(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrInvalidInput)

	_, err := handler.Register(ctx, &userv1.RegisterRequest{})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestRegisterHandler_AlreadyExists(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().Register(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrEmailAlreadyExists)

	_, err := handler.Register(ctx, &userv1.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		FullName: "Existing",
	})

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.AlreadyExists {
		t.Errorf("expected AlreadyExists, got %v", st.Code())
	}
}

func TestLoginHandler_Success(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	tokenPair := domain.TokenPair{
		AccessToken:           "access-jwt",
		AccessTokenExpiresAt:  time.Now().Add(60 * time.Minute),
		RefreshToken:          "refresh-jwt",
		RefreshTokenExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	uc.EXPECT().Login(ctx, "test@example.com", "password123").Return(domain.User{
		ID:    "user-123",
		Email: "test@example.com",
	}, tokenPair, nil)

	resp, err := handler.Login(ctx, &userv1.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AuthToken.AccessToken != "access-jwt" {
		t.Errorf("expected access-jwt, got %s", resp.AuthToken.AccessToken)
	}
	if resp.AuthToken.RefreshToken != "refresh-jwt" {
		t.Errorf("expected refresh-jwt, got %s", resp.AuthToken.RefreshToken)
	}
	if resp.AuthToken.AccessTokenExpiresAtUnixMs == 0 {
		t.Error("expected access token expiry to be set")
	}
	if resp.AuthToken.RefreshTokenExpiresAtUnixMs == 0 {
		t.Error("expected refresh token expiry to be set")
	}
}

func TestLoginHandler_Unauthenticated(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().Login(ctx, gomock.Any(), gomock.Any()).Return(domain.User{}, domain.TokenPair{}, domain.ErrInvalidCredential)

	_, err := handler.Login(ctx, &userv1.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong",
	})

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestLoginHandler_UserDisabled(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().Login(ctx, gomock.Any(), gomock.Any()).Return(domain.User{}, domain.TokenPair{}, domain.ErrUserDisabled)

	_, err := handler.Login(ctx, &userv1.LoginRequest{
		Email:    "disabled@example.com",
		Password: "password123",
	})

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", st.Code())
	}
}

func TestGetProfileHandler_NotFound(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().GetProfile(ctx, "unknown").Return(domain.User{}, domain.ErrUserNotFound)

	_, err := handler.GetProfile(ctx, &userv1.GetProfileRequest{UserId: "unknown"})

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}
