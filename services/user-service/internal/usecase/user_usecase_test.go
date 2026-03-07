package usecase_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"stockanalyzr/pkg/validator"
	"stockanalyzr/services/user-service/internal/domain"
	"stockanalyzr/services/user-service/internal/mocks"
	"stockanalyzr/services/user-service/internal/usecase"
)

func newTestUsecase(t *testing.T) (
	*usecase.UserUsecase,
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

	user, err := uc.Register(ctx, "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.Disabled {
		t.Error("expected user not disabled")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	uc, repo, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	repo.EXPECT().GetByEmail(ctx, "existing@example.com").Return(domain.User{Email: "existing@example.com"}, nil)

	_, err := uc.Register(ctx, "existing@example.com", "password123", "Test")
	if err != domain.ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	uc, _, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		email    string
		password string
		fullName string
	}{
		{"empty email", "", "password123", "Name"},
		{"invalid email", "notanemail", "password123", "Name"},
		{"short password", "test@example.com", "short", "Name"},
		{"empty name", "test@example.com", "password123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Register(ctx, tt.email, tt.password, tt.fullName)
			if err != domain.ErrInvalidInput {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
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
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenPair.AccessToken != "access-jwt" {
		t.Errorf("expected access-jwt, got %s", tokenPair.AccessToken)
	}
	if tokenPair.RefreshToken != "refresh-jwt" {
		t.Errorf("expected refresh-jwt, got %s", tokenPair.RefreshToken)
	}
	if user.ID != "user-123" {
		t.Errorf("expected user-123, got %s", user.ID)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	uc, repo, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	repo.EXPECT().GetByEmail(ctx, "unknown@example.com").Return(domain.User{}, domain.ErrUserNotFound)

	_, _, err := uc.Login(ctx, "unknown@example.com", "password123")
	if err != domain.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
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
	if err != domain.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
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
	if err != domain.ErrUserDisabled {
		t.Fatalf("expected ErrUserDisabled, got %v", err)
	}
}

func TestGetProfile_CacheHit(t *testing.T) {
	uc, _, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	cachedUser := domain.User{ID: "user-123", Email: "test@example.com"}
	cache.EXPECT().Get(ctx, "user-123").Return(cachedUser, nil)

	user, err := uc.GetProfile(ctx, "user-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "user-123" {
		t.Errorf("expected user-123, got %s", user.ID)
	}
}

func TestGetProfile_CacheMiss(t *testing.T) {
	uc, repo, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	dbUser := domain.User{ID: "user-123", Email: "test@example.com"}
	cache.EXPECT().Get(ctx, "user-123").Return(domain.User{}, domain.ErrUserNotFound)
	repo.EXPECT().GetByID(ctx, "user-123").Return(dbUser, nil)
	cache.EXPECT().Set(ctx, dbUser).Return(nil)

	user, err := uc.GetProfile(ctx, "user-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "user-123" {
		t.Errorf("expected user-123, got %s", user.ID)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	uc, repo, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	cache.EXPECT().Get(ctx, "unknown").Return(domain.User{}, domain.ErrUserNotFound)
	repo.EXPECT().GetByID(ctx, "unknown").Return(domain.User{}, domain.ErrUserNotFound)

	_, err := uc.GetProfile(ctx, "unknown")
	if err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	uc, repo, cache, _, _ := newTestUsecase(t)
	ctx := context.Background()

	updatedUser := domain.User{ID: "user-123", FullName: "New Name"}
	repo.EXPECT().UpdateProfile(ctx, "user-123", "New Name", gomock.Any()).Return(updatedUser, nil)
	cache.EXPECT().Delete(ctx, "user-123").Return(nil)

	user, err := uc.UpdateProfile(ctx, "user-123", "New Name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.FullName != "New Name" {
		t.Errorf("expected New Name, got %s", user.FullName)
	}
}

func TestUpdateProfile_InvalidInput(t *testing.T) {
	uc, _, _, _, _ := newTestUsecase(t)
	ctx := context.Background()

	_, err := uc.UpdateProfile(ctx, "", "Name")
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = uc.UpdateProfile(ctx, "user-123", "")
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
