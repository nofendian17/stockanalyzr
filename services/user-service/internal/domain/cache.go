package domain

import "context"

//go:generate mockgen -source=cache.go -destination=../mocks/mock_user_cache.go -package=mocks

// UserCache defines caching operations for user data.
type UserCache interface {
	Get(ctx context.Context, userID string) (User, error)
	Set(ctx context.Context, user User) error
	Delete(ctx context.Context, userID string) error
}
