package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "stockanalyzr/pkg/gen"
	"stockanalyzr/services/customer-gw-service/internal/domain"
)

// UserInteractor implements domain.UserUsecase by wrapping gRPC client calls.
type UserInteractor struct {
	client    userpb.UserServiceClient
	tokenRepo domain.TokenRepository
	jwtSecret string
}

// Compile-time interface compliance check.
var _ domain.UserUsecase = (*UserInteractor)(nil)

// NewUserUsecase creates a new UserInteractor.
func NewUserUsecase(client userpb.UserServiceClient, tokenRepo domain.TokenRepository, jwtSecret string) *UserInteractor {
	return &UserInteractor{
		client:    client,
		tokenRepo: tokenRepo,
		jwtSecret: jwtSecret,
	}
}

// Register handles user registration.
func (u *UserInteractor) Register(ctx context.Context, email, password, fullName, phoneNumber string) (domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)
	phoneNumber = strings.TrimSpace(phoneNumber)

	req := &userpb.RegisterRequest{
		Email:       email,
		Password:    password,
		FullName:    fullName,
		PhoneNumber: phoneNumber,
	}

	resp, err := u.client.Register(ctx, req)
	if err != nil {
		return domain.User{}, mapGRPCError(err)
	}

	return toDomainUser(resp.User), nil
}

// Login handles user login.
func (u *UserInteractor) Login(ctx context.Context, email, password string) (domain.User, domain.AuthToken, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	req := &userpb.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := u.client.Login(ctx, req)
	if err != nil {
		return domain.User{}, domain.AuthToken{}, mapGRPCError(err)
	}

	return toDomainUser(resp.User), toDomainAuthToken(resp.AuthToken), nil
}

// GetProfile retrieves a user profile.
func (u *UserInteractor) GetProfile(ctx context.Context, userID string) (domain.User, error) {
	userID = strings.TrimSpace(userID)

	req := &userpb.GetProfileRequest{
		UserId: userID,
	}

	resp, err := u.client.GetProfile(ctx, req)
	if err != nil {
		return domain.User{}, mapGRPCError(err)
	}

	return toDomainUser(resp.User), nil
}

// UpdateProfile updates a user profile.
func (u *UserInteractor) UpdateProfile(ctx context.Context, userID, fullName, phoneNumber string) (domain.User, error) {
	userID = strings.TrimSpace(userID)
	fullName = strings.TrimSpace(fullName)
	phoneNumber = strings.TrimSpace(phoneNumber)

	req := &userpb.UpdateProfileRequest{
		UserId:      userID,
		FullName:    fullName,
		PhoneNumber: phoneNumber,
	}

	resp, err := u.client.UpdateProfile(ctx, req)
	if err != nil {
		return domain.User{}, mapGRPCError(err)
	}

	return toDomainUser(resp.User), nil
}

// DeactivateAccount soft deletes a user account.
func (u *UserInteractor) DeactivateAccount(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)

	req := &userpb.DeactivateAccountRequest{
		UserId: userID,
	}

	_, err := u.client.DeactivateAccount(ctx, req)
	if err != nil {
		return mapGRPCError(err)
	}

	return nil
}

// Logout handles user logout by blacklisting the token.
func (u *UserInteractor) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.ErrInvalidInput
	}

	// Parse token to get expiration time for TTL calculation
	ttlSeconds, err := u.getTokenTTL(token)
	if err != nil {
		// If we can't parse the token, still try to blacklist it with a default TTL
		ttlSeconds = 3600 // 1 hour default
	}

	// If token is already expired, no need to blacklist
	if ttlSeconds <= 0 {
		return nil
	}

	// Add token to blacklist
	if err := u.tokenRepo.Blacklist(ctx, token, ttlSeconds); err != nil {
		return domain.ErrInternal
	}

	return nil
}

// getTokenTTL parses the JWT token and returns remaining TTL in seconds.
func (u *UserInteractor) getTokenTTL(tokenString string) (int, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrInvalidKey
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return 0, jwt.ErrInvalidKey
	}

	// Calculate remaining time
	expTime := time.Unix(int64(exp), 0)
	remaining := time.Until(expTime)

	if remaining <= 0 {
		return 0, nil
	}

	return int(remaining.Seconds()), nil
}

// IsTokenBlacklisted checks if a token has been blacklisted.
func (u *UserInteractor) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	return u.tokenRepo.IsBlacklisted(ctx, token)
}

// Helper functions

func toDomainUser(user *userpb.User) domain.User {
	result := domain.User{
		ID:          user.Id,
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Disabled:    user.Disabled,
		CreatedAt:   user.CreatedAtUnixMs,
		UpdatedAt:   user.UpdatedAtUnixMs,
	}

	if user.DeletedAtUnixMs != nil {
		result.DeletedAt = user.DeletedAtUnixMs
	}

	return result
}

func toDomainAuthToken(token *userpb.AuthToken) domain.AuthToken {
	return domain.AuthToken{
		AccessToken:                 token.AccessToken,
		AccessTokenExpiresAtUnixMs:  token.AccessTokenExpiresAtUnixMs,
		RefreshToken:                token.RefreshToken,
		RefreshTokenExpiresAtUnixMs: token.RefreshTokenExpiresAtUnixMs,
	}
}

func mapGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return domain.ErrInternal
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return domain.ErrInvalidInput
	case codes.NotFound:
		return domain.ErrUserNotFound
	case codes.AlreadyExists:
		return domain.ErrEmailAlreadyExists
	case codes.Unauthenticated:
		return domain.ErrInvalidCredential
	case codes.PermissionDenied:
		return domain.ErrForbidden
	case codes.FailedPrecondition:
		return domain.ErrInvalidInput
	default:
		return domain.ErrInternal
	}
}
