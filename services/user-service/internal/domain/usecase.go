package domain

import "context"

//go:generate mockgen -source=usecase.go -destination=../mocks/mock_user_usecase.go -package=mocks

// UserUsecase defines business operations for user management.
type UserUsecase interface {
	Register(ctx context.Context, email, password, fullName, phoneNumber string) (User, error)
	Login(ctx context.Context, email, password string) (User, TokenPair, error)
	GetProfile(ctx context.Context, userID string) (User, error)
	UpdateProfile(ctx context.Context, userID, fullName, phoneNumber string) (User, error)
	SoftDeleteUser(ctx context.Context, userID string) error
	RestoreUser(ctx context.Context, userID string) error
	GetDeletedUsers(ctx context.Context, limit, offset int) ([]User, error)
	Logout(ctx context.Context, token string) error
}
