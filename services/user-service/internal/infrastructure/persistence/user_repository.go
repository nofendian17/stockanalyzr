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
		INSERT INTO users (id, email, password_hash, full_name, phone_number, disabled, deleted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, email, password_hash, full_name, phone_number, disabled, deleted_at, created_at, updated_at
	`

	created := domain.User{}
	err := r.db.QueryRow(ctx, q,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.PhoneNumber,
		user.Disabled,
		user.DeletedAt,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&created.ID,
		&created.Email,
		&created.PasswordHash,
		&created.FullName,
		&created.PhoneNumber,
		&created.Disabled,
		&created.DeletedAt,
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
		SELECT id, email, password_hash, full_name, phone_number, disabled, deleted_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.PhoneNumber,
		&user.Disabled,
		&user.DeletedAt,
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
		SELECT id, email, password_hash, full_name, phone_number, disabled, deleted_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := domain.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.PhoneNumber,
		&user.Disabled,
		&user.DeletedAt,
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

func (r *UserRepository) UpdateProfile(ctx context.Context, id string, fullName string, phoneNumber string, updatedAt time.Time) (domain.User, error) {
	q := `
		UPDATE users
		SET full_name = $2, phone_number = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, full_name, phone_number, disabled, deleted_at, created_at, updated_at
	`

	updated := domain.User{}
	err := r.db.QueryRow(ctx, q, id, fullName, phoneNumber, updatedAt).Scan(
		&updated.ID,
		&updated.Email,
		&updated.PasswordHash,
		&updated.FullName,
		&updated.PhoneNumber,
		&updated.Disabled,
		&updated.DeletedAt,
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

// SoftDelete marks a user as deleted
func (r *UserRepository) SoftDelete(ctx context.Context, id string, deletedAt time.Time) error {
	q := `
		UPDATE users
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, q, id, deletedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
