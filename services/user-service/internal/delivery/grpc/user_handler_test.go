package grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "stockanalyzr/pkg/gen"
	transportgrpc "stockanalyzr/services/user-service/internal/delivery/grpc"
	"stockanalyzr/services/user-service/internal/domain"
	"stockanalyzr/services/user-service/internal/middleware"
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

	uc.EXPECT().Register(ctx, "test@example.com", "password123", "Test User", "+1234567890").Return(domain.User{
		ID:          "user-123",
		Email:       "test@example.com",
		FullName:    "Test User",
		PhoneNumber: "+1234567890",
	}, nil)

	resp, err := handler.Register(ctx, &userv1.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		FullName:    "Test User",
		PhoneNumber: "+1234567890",
	})

	require.NoError(t, err)
	assert.Equal(t, "user-123", resp.User.Id)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "Test User", resp.User.FullName)
	assert.Equal(t, "+1234567890", resp.User.PhoneNumber)
}

func TestRegisterHandler_InvalidInput(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().Register(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrInvalidInput)

	_, err := handler.Register(ctx, &userv1.RegisterRequest{})

	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestRegisterHandler_AlreadyExists(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().Register(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrEmailAlreadyExists)

	_, err := handler.Register(ctx, &userv1.RegisterRequest{
		Email:       "existing@example.com",
		Password:    "password123",
		FullName:    "Existing",
		PhoneNumber: "+123",
	})

	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error")
	assert.Equal(t, codes.AlreadyExists, st.Code())
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

	require.NoError(t, err)
	assert.Equal(t, "access-jwt", resp.AuthToken.AccessToken)
	assert.Equal(t, "refresh-jwt", resp.AuthToken.RefreshToken)
	assert.NotZero(t, resp.AuthToken.AccessTokenExpiresAtUnixMs)
	assert.NotZero(t, resp.AuthToken.RefreshTokenExpiresAtUnixMs)
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
	require.True(t, ok, "expected gRPC status error")
	assert.Equal(t, codes.Unauthenticated, st.Code())
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
	require.True(t, ok, "expected gRPC status error")
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestGetProfileHandler_NotFound(t *testing.T) {
	handler, uc := newTestHandler(t)
	ctx := context.Background()

	uc.EXPECT().GetProfile(ctx, "unknown").Return(domain.User{}, domain.ErrUserNotFound)

	_, err := handler.GetProfile(ctx, &userv1.GetProfileRequest{UserId: "unknown"})

	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error")
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestLogoutHandler_Success(t *testing.T) {
	handler, uc := newTestHandler(t)

	// Create context with token using the new middleware helper
	ctx := middleware.ContextWithAuthData(context.Background(), middleware.AuthData{
		AccessToken: "valid-token",
	})

	uc.EXPECT().Logout(gomock.Any(), "valid-token").Return(nil)

	resp, err := handler.Logout(ctx, &userv1.LogoutRequest{})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestLogoutHandler_MissingToken(t *testing.T) {
	handler, _ := newTestHandler(t)

	// Context without token
	ctx := context.Background()

	_, err := handler.Logout(ctx, &userv1.LogoutRequest{})

	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error")
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestUpdateProfileHandler_Success(t *testing.T) {
	handler, uc := newTestHandler(t)

	ctx := middleware.ContextWithAuthData(context.Background(), middleware.AuthData{
		UserID:      "user-123",
		AccessToken: "valid-token",
	})

	uc.EXPECT().UpdateProfile(ctx, "user-123", "New Name", "+1234").Return(domain.User{
		ID:          "user-123",
		FullName:    "New Name",
		PhoneNumber: "+1234",
	}, nil)

	resp, err := handler.UpdateProfile(ctx, &userv1.UpdateProfileRequest{
		UserId:      "user-123",
		FullName:    "New Name",
		PhoneNumber: "+1234",
	})

	require.NoError(t, err)
	assert.Equal(t, "user-123", resp.User.Id)
	assert.Equal(t, "New Name", resp.User.FullName)
	assert.Equal(t, "+1234", resp.User.PhoneNumber)
}

func TestUpdateProfileHandler_Unauthorized(t *testing.T) {
	handler, _ := newTestHandler(t)

	// User logged in as user-123, tries to update user-456
	ctx := middleware.ContextWithAuthData(context.Background(), middleware.AuthData{
		UserID:      "user-123",
		AccessToken: "valid-token",
	})

	_, err := handler.UpdateProfile(ctx, &userv1.UpdateProfileRequest{
		UserId:      "user-456",
		FullName:    "Hacked Name",
		PhoneNumber: "000",
	})

	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error")
	assert.Equal(t, codes.PermissionDenied, st.Code())
}
