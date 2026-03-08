package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"stockanalyzr/pkg/validator"
	"stockanalyzr/services/user-service/internal/domain"
	"stockanalyzr/services/user-service/internal/mocks"
	"stockanalyzr/services/user-service/internal/usecase"
)

func newTestUsecase(t *testing.T) (
	*usecase.UserInteractor,
	*mocks.MockUserRepository,
	*mocks.MockUserCache,
	*mocks.MockPasswordHasher,
	*mocks.MockTokenManager,
) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	cache := mocks.NewMockUserCache(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	tokenMgr := mocks.NewMockTokenManager(ctrl)

	val := validator.New()

	uc := usecase.NewUserUsecase(repo, cache, hasher, tokenMgr, val)
	return uc, repo, cache, hasher, tokenMgr
}

func TestRegister_Success(t *testing.T) {
	uc, repo, cache, hasher, _ := newTestUsecase(t)
	ctx := context.Background()

	repo.EXPECT().GetByEmail(ctx, "test@example.com").Return(domain.User{}, domain.ErrUserNotFound)
	hasher.EXPECT().Hash("password123").Return("hashed", nil)
	repo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, u domain.User) (domain.User, error) {
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
		return u, nil
	})
	cache.EXPECT().Set(ctx, gomock.Any()).Return(nil)

	user, err := uc.Register(ctx, "test@example.com", "password123", "Test User", "+1234567890")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "+1234567890", user.PhoneNumber)
	assert.False(t, user.Disabled)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	uc, repo, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	repo.EXPECT().GetByEmail(ctx, "existing@example.com").Return(domain.User{Email: "existing@example.com"}, nil)

	_, err := uc.Register(ctx, "existing@example.com", "password123", "Test", "+123")
	require.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
}

func TestRegister_InvalidInput(t *testing.T) {
	uc, _, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		email       string
		password    string
		fullName    string
		phoneNumber string
	}{
		{"empty email", "", "password123", "Name", "+123"},
		{"invalid email", "notanemail", "password123", "Name", "+123"},
		{"short password", "test@example.com", "short", "Name", "+123"},
		{"empty name", "test@example.com", "password123", "", "+123"},
		{"empty phone", "test@example.com", "password123", "Name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Register(ctx, tt.email, tt.password, tt.fullName, tt.phoneNumber)
			require.ErrorIs(t, err, domain.ErrInvalidInput)
		})
	}
}

func TestLogin_Success(t *testing.T) {
	uc, repo, _, hasher, tokenMgr := newTestUsecase(t)
	ctx := context.Background()

	existingUser := domain.User{
		ID:           "user-123",
		Email:        "test@example.com",
		PasswordHash: "hashed",
		FullName:     "Test User",
		Disabled:     false,
	}

	expectedTokenPair := domain.TokenPair{
		AccessToken:           "access-jwt",
		AccessTokenExpiresAt:  time.Now().Add(60 * time.Minute),
		RefreshToken:          "refresh-jwt",
		RefreshTokenExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	repo.EXPECT().GetByEmail(ctx, "test@example.com").Return(existingUser, nil)
	hasher.EXPECT().Compare("hashed", "password123").Return(nil)
	tokenMgr.EXPECT().CreateTokenPair(ctx, "user-123").Return(expectedTokenPair, nil)

	user, tokenPair, err := uc.Login(ctx, "test@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, "access-jwt", tokenPair.AccessToken)
	assert.Equal(t, "refresh-jwt", tokenPair.RefreshToken)
	assert.Equal(t, "user-123", user.ID)
}

func TestLogin_UserNotFound(t *testing.T) {
	uc, repo, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	repo.EXPECT().GetByEmail(ctx, "unknown@example.com").Return(domain.User{}, domain.ErrUserNotFound)

	_, _, err := uc.Login(ctx, "unknown@example.com", "password123")
	require.ErrorIs(t, err, domain.ErrInvalidCredential)
}

func TestLogin_WrongPassword(t *testing.T) {
	uc, repo, _, hasher, _ := newTestUsecase(t)
	ctx := context.Background()

	existingUser := domain.User{
		ID:           "user-123",
		Email:        "test@example.com",
		PasswordHash: "hashed",
	}

	repo.EXPECT().GetByEmail(ctx, "test@example.com").Return(existingUser, nil)
	hasher.EXPECT().Compare("hashed", "wrongpassword").Return(domain.ErrInvalidCredential)

	_, _, err := uc.Login(ctx, "test@example.com", "wrongpassword")
	require.ErrorIs(t, err, domain.ErrInvalidCredential)
}

func TestLogin_UserDisabled(t *testing.T) {
	uc, repo, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	disabledUser := domain.User{
		ID:       "user-123",
		Email:    "disabled@example.com",
		Disabled: true,
	}

	repo.EXPECT().GetByEmail(ctx, "disabled@example.com").Return(disabledUser, nil)

	_, _, err := uc.Login(ctx, "disabled@example.com", "password123")
	require.ErrorIs(t, err, domain.ErrUserDisabled)
}

func TestGetProfile_CacheHit(t *testing.T) {
	uc, _, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	cachedUser := domain.User{ID: "user-123", Email: "test@example.com"}
	cache.EXPECT().Get(ctx, "user-123").Return(cachedUser, nil)

	user, err := uc.GetProfile(ctx, "user-123")
	require.NoError(t, err)
	assert.Equal(t, "user-123", user.ID)
}

func TestGetProfile_CacheMiss(t *testing.T) {
	uc, repo, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	dbUser := domain.User{ID: "user-123", Email: "test@example.com"}
	cache.EXPECT().Get(ctx, "user-123").Return(domain.User{}, domain.ErrUserNotFound)
	repo.EXPECT().GetByID(ctx, "user-123").Return(dbUser, nil)
	cache.EXPECT().Set(ctx, dbUser).Return(nil)

	user, err := uc.GetProfile(ctx, "user-123")
	require.NoError(t, err)
	assert.Equal(t, "user-123", user.ID)
}

func TestGetProfile_NotFound(t *testing.T) {
	uc, repo, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	cache.EXPECT().Get(ctx, "unknown").Return(domain.User{}, domain.ErrUserNotFound)
	repo.EXPECT().GetByID(ctx, "unknown").Return(domain.User{}, domain.ErrUserNotFound)

	_, err := uc.GetProfile(ctx, "unknown")
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUpdateProfile_Success(t *testing.T) {
	uc, repo, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	updatedUser := domain.User{ID: "user-123", FullName: "New Name", PhoneNumber: "+1234"}
	repo.EXPECT().UpdateProfile(ctx, "user-123", "New Name", "+1234", gomock.Any()).Return(updatedUser, nil)
	cache.EXPECT().Delete(ctx, "user-123").Return(nil)

	user, err := uc.UpdateProfile(ctx, "user-123", "New Name", "+1234")
	require.NoError(t, err)
	assert.Equal(t, "New Name", user.FullName)
	assert.Equal(t, "+1234", user.PhoneNumber)
}

func TestUpdateProfile_InvalidInput(t *testing.T) {
	uc, _, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	_, err := uc.UpdateProfile(ctx, "", "Name", "+123")
	require.ErrorIs(t, err, domain.ErrInvalidInput)

	_, err = uc.UpdateProfile(ctx, "user-123", "", "+123")
	require.ErrorIs(t, err, domain.ErrInvalidInput)

	_, err = uc.UpdateProfile(ctx, "user-123", "Name", "")
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestLogout_Success(t *testing.T) {
	uc, _, _, _, tokenMgr := newTestUsecase(t)
	ctx := context.Background()

	tokenMgr.EXPECT().RevokeToken(ctx, "valid-token").Return(nil)

	err := uc.Logout(ctx, "valid-token")
	require.NoError(t, err)
}

func TestLogout_InvalidInput(t *testing.T) {
	uc, _, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	err := uc.Logout(ctx, "   ")
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestSoftDeleteUser_Success(t *testing.T) {
	uc, repo, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()
	userID := "user123"

	// Mock existing user (not deleted)
	existingUser := domain.User{
		ID:        userID,
		Email:     "test@example.com",
		DeletedAt: nil,
	}

	repo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)
	repo.EXPECT().SoftDelete(ctx, userID, gomock.Any()).Return(nil)
	cache.EXPECT().Delete(ctx, userID).Return(nil)

	err := uc.SoftDeleteUser(ctx, userID)
	assert.NoError(t, err)
}

func TestSoftDeleteUser_AlreadyDeleted(t *testing.T) {
	uc, repo, _, _, _ := newTestUsecase(t)
	ctx := context.Background()
	userID := "user123"
	deletedAt := time.Now().UTC()

	// Mock existing user (already deleted)
	existingUser := domain.User{
		ID:        userID,
		Email:     "test@example.com",
		DeletedAt: &deletedAt,
	}

	repo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)

	err := uc.SoftDeleteUser(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserAlreadyDeleted, err)
}

func TestSoftDeleteUser_UserNotFound(t *testing.T) {
	uc, repo, _, _, _ := newTestUsecase(t)
	ctx := context.Background()
	userID := "user123"

	repo.EXPECT().GetByID(ctx, userID).Return(domain.User{}, domain.ErrUserNotFound)

	err := uc.SoftDeleteUser(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
}
