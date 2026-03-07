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
	Email       string `validate:"required,email"`
	Password    string `validate:"required,min=8"`
	FullName    string `validate:"required"`
	PhoneNumber string `validate:"required"`
}

type loginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type profileInput struct {
	UserID string `validate:"required"`
}

type updateProfileInput struct {
	UserID      string `validate:"required"`
	FullName    string `validate:"required"`
	PhoneNumber string `validate:"required"`
}

type softDeleteInput struct {
	UserID string `validate:"required"`
}

type restoreInput struct {
	UserID string `validate:"required"`
}

type getDeletedInput struct {
	Limit  int `validate:"min=1,max=100"`
	Offset int `validate:"min=0"`
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

func (u *UserUsecase) Register(ctx context.Context, email, password, fullName, phoneNumber string) (domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)
	phoneNumber = strings.TrimSpace(phoneNumber)

	if err := u.validator.Struct(registerInput{
		Email:       email,
		Password:    password,
		FullName:    fullName,
		PhoneNumber: phoneNumber,
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
		PhoneNumber:  phoneNumber,
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

	// Check if user is deleted
	if user.IsDeleted() {
		return domain.User{}, domain.TokenPair{}, domain.ErrUserNotFound
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

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID, fullName, phoneNumber string) (domain.User, error) {
	userID = strings.TrimSpace(userID)
	fullName = strings.TrimSpace(fullName)
	phoneNumber = strings.TrimSpace(phoneNumber)

	if err := u.validator.Struct(updateProfileInput{UserID: userID, FullName: fullName, PhoneNumber: phoneNumber}); err != nil {
		return domain.User{}, domain.ErrInvalidInput
	}

	updated, err := u.repo.UpdateProfile(ctx, userID, fullName, phoneNumber, time.Now().UTC())
	if err != nil {
		return domain.User{}, err
	}

	// Invalidate cache (best effort)
	if u.cache != nil {
		_ = u.cache.Delete(ctx, userID)
	}

	return updated, nil
}

func (u *UserUsecase) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.ErrInvalidInput
	}

	// Delegate revocation to TokenManager
	return u.tokenManager.RevokeToken(ctx, token)
}

// SoftDeleteUser soft deletes a user account
func (u *UserUsecase) SoftDeleteUser(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if err := u.validator.Struct(softDeleteInput{UserID: userID}); err != nil {
		return domain.ErrInvalidInput
	}

	// Check if user exists and is not already deleted
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.IsDeleted() {
		return domain.ErrUserAlreadyDeleted
	}

	// Perform soft delete
	deletedAt := time.Now().UTC()
	if err := u.repo.SoftDelete(ctx, userID, deletedAt); err != nil {
		return err
	}

	// Invalidate cache
	if u.cache != nil {
		_ = u.cache.Delete(ctx, userID)
	}

	return nil
}

// RestoreUser restores a soft deleted user
func (u *UserUsecase) RestoreUser(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if err := u.validator.Struct(restoreInput{UserID: userID}); err != nil {
		return domain.ErrInvalidInput
	}

	// Check if user exists and is deleted
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.IsDeleted() {
		return domain.ErrUserNotDeleted
	}

	// Perform restore
	if err := u.repo.Restore(ctx, userID); err != nil {
		return err
	}

	// Invalidate cache
	if u.cache != nil {
		_ = u.cache.Delete(ctx, userID)
	}

	return nil
}

// GetDeletedUsers retrieves paginated list of deleted users
func (u *UserUsecase) GetDeletedUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	if err := u.validator.Struct(getDeletedInput{Limit: limit, Offset: offset}); err != nil {
		return nil, domain.ErrInvalidInput
	}

	return u.repo.GetDeletedUsers(ctx, limit, offset)
}
