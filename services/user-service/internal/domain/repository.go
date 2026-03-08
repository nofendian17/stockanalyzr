package domain

import (
	"context"
	"time"
)

//go:generate mockgen -source=repository.go -destination=../mocks/mock_user_repository.go -package=mocks

// UserRepository defines persistence boundary.
type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	UpdateProfile(ctx context.Context, id string, fullName string, phoneNumber string, updatedAt time.Time) (User, error)
	SoftDelete(ctx context.Context, id string, deletedAt time.Time) error
}
