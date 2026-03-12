package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"

	"stockanalyzr/pkg/response"
	"stockanalyzr/services/customer-gw-service/internal/delivery/dto"
	"stockanalyzr/services/customer-gw-service/internal/domain"
	"stockanalyzr/services/customer-gw-service/internal/middleware"
)

// withAuthMetadata creates a context with authorization metadata from the request.
func withAuthMetadata(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		md := metadata.New(map[string]string{
			"authorization": authHeader,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	uc domain.UserUsecase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{
		uc: uc,
	}
}

// Register handles user registration.
//
//	@Summary		Register a new user
//	@Description	Create a new user account with email, password, and profile details
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"Registration details"
//	@Success		201		{object}	response.Response{data=dto.RegisterResponse}	"User registered successfully"
//	@Failure		400		{object}	response.Response	"Invalid input"
//	@Failure		409		{object}	response.Response	"Email already exists"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, err := h.uc.Register(c.Request.Context(), req.Email, req.Password, req.FullName, req.PhoneNumber)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	data := dto.RegisterResponse{
		User: toUserResponse(user),
	}
	response.Success(c, http.StatusCreated, data)
}

// Login handles user login.
//
//	@Summary		User login
//	@Description	Authenticate user and return access/refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	response.Response{data=dto.LoginResponse}	"Login successful"
//	@Failure		400		{object}	response.Response	"Invalid input"
//	@Failure		401		{object}	response.Response	"Invalid credentials"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, token, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	data := dto.LoginResponse{
		User: toUserResponse(user),
		AuthToken: dto.AuthTokenResponse{
			AccessToken:                 token.AccessToken,
			AccessTokenExpiresAtUnixMs:  token.AccessTokenExpiresAtUnixMs,
			RefreshToken:                token.RefreshToken,
			RefreshTokenExpiresAtUnixMs: token.RefreshTokenExpiresAtUnixMs,
		},
	}
	response.Success(c, http.StatusOK, data)
}

// RefreshToken handles token refresh.
//
//	@Summary		Refresh access token
//	@Description	Refresh access token using a valid refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshTokenRequest	true	"Refresh token"
//	@Success		200		{object}	response.Response{data=dto.RefreshTokenResponse}	"Token refreshed successfully"
//	@Failure		400		{object}	response.Response	"Invalid input"
//	@Failure		401		{object}	response.Response	"Invalid or expired refresh token"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/api/v1/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	token, err := h.uc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	data := dto.RefreshTokenResponse{
		AuthToken: dto.AuthTokenResponse{
			AccessToken:                 token.AccessToken,
			AccessTokenExpiresAtUnixMs:  token.AccessTokenExpiresAtUnixMs,
			RefreshToken:                token.RefreshToken,
			RefreshTokenExpiresAtUnixMs: token.RefreshTokenExpiresAtUnixMs,
		},
	}
	response.Success(c, http.StatusOK, data)
}

// GetProfile handles getting the authenticated user's profile.
//
//	@Summary		Get user profile
//	@Description	Retrieve the authenticated user's profile
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=dto.ProfileResponse}	"Profile retrieved successfully"
//	@Failure		401	{object}	response.Response	"Unauthorized"
//	@Failure		404	{object}	response.Response	"User not found"
//	@Router			/api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}

	user, err := h.uc.GetProfile(withAuthMetadata(c), authUser.UserID)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	data := dto.ProfileResponse{
		User: toUserResponse(user),
	}
	response.Success(c, http.StatusOK, data)
}

// UpdateProfile handles updating the authenticated user's profile.
//
//	@Summary		Update user profile
//	@Description	Update the authenticated user's profile information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.UpdateProfileRequest	true	"Profile update details"
//	@Success		200		{object}	response.Response{data=dto.UpdateProfileResponse}	"Profile updated successfully"
//	@Failure		400		{object}	response.Response	"Invalid input"
//	@Failure		401		{object}	response.Response	"Unauthorized"
//	@Failure		404		{object}	response.Response	"User not found"
//	@Router			/api/v1/users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, err := h.uc.UpdateProfile(withAuthMetadata(c), authUser.UserID, req.FullName, req.PhoneNumber)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	data := dto.UpdateProfileResponse{
		User: toUserResponse(user),
	}
	response.Success(c, http.StatusOK, data)
}

// Logout handles user logout.
//
//	@Summary		User logout
//	@Description	Invalidate the current access token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=dto.LogoutResponse}	"Logged out successfully"
//	@Failure		401	{object}	response.Response	"Unauthorized"
//	@Router			/api/v1/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "logged out successfully", dto.LogoutResponse{
		Success: true,
	})
}

// DeactivateAccount handles deactivating the authenticated user's account.
//
//	@Summary		Deactivate user account
//	@Description	Deactivate the authenticated user's account
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=dto.DeactivateResponse}	"Account deactivated successfully"
//	@Failure		401	{object}	response.Response	"Unauthorized"
//	@Failure		404	{object}	response.Response	"User not found"
//	@Router			/api/v1/users/me [delete]
func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		response.Unauthorized(c, "")
		return
	}

	if err := h.uc.DeactivateAccount(withAuthMetadata(c), authUser.UserID); err != nil {
		h.handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.DeactivateResponse{
		Success: true,
	})
}

// toUserResponse converts a domain User to a UserResponse DTO.
func toUserResponse(user domain.User) dto.UserResponse {
	result := dto.UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		FullName:        user.FullName,
		PhoneNumber:     user.PhoneNumber,
		Disabled:        user.Disabled,
		CreatedAtUnixMs: user.CreatedAt,
		UpdatedAtUnixMs: user.UpdatedAt,
	}

	if user.DeletedAt != nil {
		result.DeletedAtUnixMs = user.DeletedAt
	}

	return result
}

// handleDomainError converts domain errors to HTTP responses.
func (h *UserHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		response.BadRequest(c, "invalid input")
	case errors.Is(err, domain.ErrUserNotFound):
		response.NotFound(c, "user not found")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		response.Error(c, http.StatusConflict, "email already exists")
	case errors.Is(err, domain.ErrInvalidCredential):
		response.Unauthorized(c, "invalid credentials")
	case errors.Is(err, domain.ErrUserDisabled):
		response.Error(c, http.StatusForbidden, "user is disabled")
	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(c, "unauthorized")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(c, "forbidden")
	default:
		response.InternalServerError(c, "internal server error")
	}
}
