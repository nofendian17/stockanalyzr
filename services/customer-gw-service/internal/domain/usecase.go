package domain

import "context"

// UserUsecase defines business operations for user management via gRPC.
type UserUsecase interface {
	Register(ctx context.Context, email, password, fullName, phoneNumber string) (User, error)
	Login(ctx context.Context, email, password string) (User, AuthToken, error)
	GetProfile(ctx context.Context, userID string) (User, error)
	UpdateProfile(ctx context.Context, userID, fullName, phoneNumber string) (User, error)
	DeactivateAccount(ctx context.Context, userID string) error
	Logout(ctx context.Context, token string) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}
