package grpc

import (
	"context"

	"google.golang.org/grpc/codes"

	userv1 "stockanalyzr/pkg/gen"
	"stockanalyzr/pkg/grpcerr"
	"stockanalyzr/services/user-service/internal/domain"
	"stockanalyzr/services/user-service/internal/middleware"
)

// UserHandler adapts gRPC contract to usecase layer.
type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	uc domain.UserUsecase
}

func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	user, err := h.uc.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetFullName())
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.RegisterResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	user, tokenPair, err := h.uc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.LoginResponse{
		User: toProtoUser(user),
		AuthToken: &userv1.AuthToken{
			AccessToken:                 tokenPair.AccessToken,
			AccessTokenExpiresAtUnixMs:  tokenPair.AccessTokenExpiresAt.UnixMilli(),
			RefreshToken:                tokenPair.RefreshToken,
			RefreshTokenExpiresAtUnixMs: tokenPair.RefreshTokenExpiresAt.UnixMilli(),
		},
	}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	userID := req.GetUserId()
	if authData, ok := middleware.AuthDataFromContext(ctx); ok {
		userID = authData.UserID
	}

	user, err := h.uc.GetProfile(ctx, userID)
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.GetProfileResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	userID := req.GetUserId()
	if authData, ok := middleware.AuthDataFromContext(ctx); ok {
		userID = authData.UserID
	}

	user, err := h.uc.UpdateProfile(ctx, userID, req.GetFullName())
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.UpdateProfileResponse{User: toProtoUser(user)}, nil
}

func toProtoUser(user domain.User) *userv1.User {
	return &userv1.User{
		Id:              user.ID,
		Email:           user.Email,
		FullName:        user.FullName,
		Disabled:        user.Disabled,
		CreatedAtUnixMs: user.CreatedAt.UnixMilli(),
		UpdatedAtUnixMs: user.UpdatedAt.UnixMilli(),
	}
}

var userErrorMap = grpcerr.ErrorMap{
	domain.ErrInvalidInput:       codes.InvalidArgument,
	domain.ErrUserNotFound:       codes.NotFound,
	domain.ErrInvalidCredential:  codes.Unauthenticated,
	domain.ErrEmailAlreadyExists: codes.AlreadyExists,
	domain.ErrUserDisabled:       codes.PermissionDenied,
}

func toStatusErr(err error) error {
	return grpcerr.Handle(err, userErrorMap)
}
