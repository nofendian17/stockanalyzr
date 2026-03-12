package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	user, err := h.uc.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetFullName(), req.GetPhoneNumber())
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

func (h *UserHandler) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	tokenPair, err := h.uc.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.RefreshTokenResponse{
		AuthToken: &userv1.AuthToken{
			AccessToken:                 tokenPair.AccessToken,
			AccessTokenExpiresAtUnixMs:  tokenPair.AccessTokenExpiresAt.UnixMilli(),
			RefreshToken:                tokenPair.RefreshToken,
			RefreshTokenExpiresAtUnixMs: tokenPair.RefreshTokenExpiresAt.UnixMilli(),
		},
	}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.uc.GetProfile(ctx, req.GetUserId())
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.GetProfileResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	token := ""
	if authData, ok := middleware.AuthDataFromContext(ctx); ok {
		token = authData.AccessToken
	}

	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	err := h.uc.Logout(ctx, token)
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.LogoutResponse{Success: true}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	userID := req.GetUserId()
	if authData, ok := middleware.AuthDataFromContext(ctx); ok {
		// Enforce user can only update their own profile
		if authData.UserID != userID {
			return nil, status.Error(codes.PermissionDenied, "unauthorized profile update")
		}
	}

	user, err := h.uc.UpdateProfile(ctx, userID, req.GetFullName(), req.GetPhoneNumber())
	if err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.UpdateProfileResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) DeactivateAccount(ctx context.Context, req *userv1.DeactivateAccountRequest) (*userv1.DeactivateAccountResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Authorization: users can only delete their own account
	if authData, ok := middleware.AuthDataFromContext(ctx); ok {
		if authData.UserID != userID {
			return nil, status.Error(codes.PermissionDenied, "unauthorized account deletion")
		}
	} else {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if err := h.uc.DeactivateAccount(ctx, userID); err != nil {
		return nil, toStatusErr(err)
	}

	return &userv1.DeactivateAccountResponse{Success: true}, nil
}

func toProtoUser(user domain.User) *userv1.User {
	protoUser := &userv1.User{
		Id:              user.ID,
		Email:           user.Email,
		FullName:        user.FullName,
		PhoneNumber:     user.PhoneNumber,
		Disabled:        user.Disabled,
		CreatedAtUnixMs: user.CreatedAt.UnixMilli(),
		UpdatedAtUnixMs: user.UpdatedAt.UnixMilli(),
	}

	if user.DeletedAt != nil {
		deletedAtUnixMs := user.DeletedAt.UnixMilli()
		protoUser.DeletedAtUnixMs = &deletedAtUnixMs
	}

	return protoUser
}

var userErrorMap = grpcerr.ErrorMap{
	domain.ErrInvalidInput:       codes.InvalidArgument,
	domain.ErrUserNotFound:       codes.NotFound,
	domain.ErrInvalidCredential:  codes.Unauthenticated,
	domain.ErrUnauthorized:       codes.Unauthenticated,
	domain.ErrEmailAlreadyExists: codes.AlreadyExists,
	domain.ErrUserDisabled:       codes.PermissionDenied,
	domain.ErrUserAlreadyDeleted: codes.FailedPrecondition,
	domain.ErrUserNotDeleted:     codes.FailedPrecondition,
}

func toStatusErr(err error) error {
	return grpcerr.Handle(err, userErrorMap)
}
