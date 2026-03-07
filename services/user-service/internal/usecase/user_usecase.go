package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"stockanalyzr/pkg/validator"
	"stockanalyzr/services/user-service/internal/domain"
)

// UserUsecase contains business rules for user operations.
type UserUsecase struct {
	repo         domain.UserRepository
	cache        domain.UserCache
	hasher       domain.PasswordHasher
	tokenManager domain.TokenManager
	validator    *validator.Validator
}

// Compile-time interface compliance check.
var _ domain.UserUsecase = (*UserUsecase)(nil)

type registerInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	FullName string `validate:"required"`
}

type loginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type profileInput struct {
	UserID string `validate:"required"`
}

type updateProfileInput struct {
	UserID   string `validate:"required"`
	FullName string `validate:"required"`
}

func NewUserUsecase(
	repo domain.UserRepository,
	cache domain.UserCache,
	hasher domain.PasswordHasher,
	tokenManager domain.TokenManager,
	val *validator.Validator,
) *UserUsecase {
	return &UserUsecase{
		repo:         repo,
		cache:        cache,
		hasher:       hasher,
		tokenManager: tokenManager,
		validator:    val,
	}
}

func (u *UserUsecase) Register(ctx context.Context, email, password, fullName string) (domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)

	if err := u.validator.Struct(registerInput{
		Email:    email,
		Password: password,
		FullName: fullName,
	}); err != nil {
		return domain.User{}, domain.ErrInvalidInput
	}

	_, err := u.repo.GetByEmail(ctx, email)
	if err == nil {
		return domain.User{}, domain.ErrEmailAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, err
	}

	hashedPassword, err := u.hasher.Hash(password)
	if err != nil {
		return domain.User{}, err
	}

	now := time.Now().UTC()
	created, err := u.repo.Create(ctx, domain.User{
		ID:           ulid.Make().String(),
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		Disabled:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return domain.User{}, err
	}

	// Populate cache (best effort)
	if u.cache != nil {
		_ = u.cache.Set(ctx, created)
	}

	return created, nil
}

func (u *UserUsecase) Login(ctx context.Context, email, password string) (domain.User, domain.TokenPair, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := u.validator.Struct(loginInput{Email: email, Password: password}); err != nil {
		return domain.User{}, domain.TokenPair{}, domain.ErrInvalidInput
	}

	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, domain.TokenPair{}, domain.ErrInvalidCredential
		}
		return domain.User{}, domain.TokenPair{}, err
	}

	if user.Disabled {
		return domain.User{}, domain.TokenPair{}, domain.ErrUserDisabled
	}

	if err := u.hasher.Compare(user.PasswordHash, password); err != nil {
		return domain.User{}, domain.TokenPair{}, domain.ErrInvalidCredential
	}

	tokenPair, err := u.tokenManager.CreateTokenPair(ctx, user.ID)
	if err != nil {
		return domain.User{}, domain.TokenPair{}, err
	}

	return user, tokenPair, nil
}

func (u *UserUsecase) GetProfile(ctx context.Context, userID string) (domain.User, error) {
	userID = strings.TrimSpace(userID)
	if err := u.validator.Struct(profileInput{UserID: userID}); err != nil {
		return domain.User{}, domain.ErrInvalidInput
	}

	// Try cache first
	if u.cache != nil {
		user, err := u.cache.Get(ctx, userID)
		if err == nil {
			return user, nil
		}
	}

	// Fallback to repository
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	// Populate cache (best effort)
	if u.cache != nil {
		_ = u.cache.Set(ctx, user)
	}

	return user, nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID, fullName string) (domain.User, error) {
	userID = strings.TrimSpace(userID)
	fullName = strings.TrimSpace(fullName)

	if err := u.validator.Struct(updateProfileInput{UserID: userID, FullName: fullName}); err != nil {
		return domain.User{}, domain.ErrInvalidInput
	}

	updated, err := u.repo.UpdateProfile(ctx, userID, fullName, time.Now().UTC())
	if err != nil {
		return domain.User{}, err
	}

	// Invalidate cache (best effort)
	if u.cache != nil {
		_ = u.cache.Delete(ctx, userID)
	}

	return updated, nil
}
