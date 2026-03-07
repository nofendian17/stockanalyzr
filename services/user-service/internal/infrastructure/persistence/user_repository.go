package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stockanalyzr/services/user-service/internal/domain"
)

// UserRepository is PostgreSQL implementation of domain.UserRepository.
type UserRepository struct {
	db *pgxpool.Pool
}

// Compile-time interface compliance check.
var _ domain.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	q := `
		INSERT INTO users (id, email, password_hash, full_name, disabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, password_hash, full_name, disabled, created_at, updated_at
	`

	created := domain.User{}
	err := r.db.QueryRow(ctx, q,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Disabled,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&created.ID,
		&created.Email,
		&created.PasswordHash,
		&created.FullName,
		&created.Disabled,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	return created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	q := `
		SELECT id, email, password_hash, full_name, disabled, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Disabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	q := `
		SELECT id, email, password_hash, full_name, disabled, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := domain.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Disabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id string, fullName string, updatedAt time.Time) (domain.User, error) {
	q := `
		UPDATE users
		SET full_name = $2, updated_at = $3
		WHERE id = $1
		RETURNING id, email, password_hash, full_name, disabled, created_at, updated_at
	`

	updated := domain.User{}
	err := r.db.QueryRow(ctx, q, id, fullName, updatedAt).Scan(
		&updated.ID,
		&updated.Email,
		&updated.PasswordHash,
		&updated.FullName,
		&updated.Disabled,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return updated, nil
}
