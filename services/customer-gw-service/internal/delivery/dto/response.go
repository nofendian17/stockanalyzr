package dto

// UserResponse represents the user data in HTTP response.
type UserResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	FullName        string `json:"full_name"`
	PhoneNumber     string `json:"phone_number"`
	Disabled        bool   `json:"disabled"`
	CreatedAtUnixMs int64  `json:"created_at_unix_ms"`
	UpdatedAtUnixMs int64  `json:"updated_at_unix_ms"`
	DeletedAtUnixMs *int64 `json:"deleted_at_unix_ms,omitempty"`
}

// AuthTokenResponse represents the authentication token in HTTP response.
type AuthTokenResponse struct {
	AccessToken                 string `json:"access_token"`
	AccessTokenExpiresAtUnixMs  int64  `json:"access_token_expires_at_unix_ms"`
	RefreshToken                string `json:"refresh_token"`
	RefreshTokenExpiresAtUnixMs int64  `json:"refresh_token_expires_at_unix_ms"`
}

// LoginResponse represents the login response data.
type LoginResponse struct {
	User      UserResponse      `json:"user"`
	AuthToken AuthTokenResponse `json:"auth_token"`
}

// RegisterResponse represents the register response data.
type RegisterResponse struct {
	User UserResponse `json:"user"`
}

// ProfileResponse represents the profile response data.
type ProfileResponse struct {
	User UserResponse `json:"user"`
}

// UpdateProfileResponse represents the update profile response data.
type UpdateProfileResponse struct {
	User UserResponse `json:"user"`
}

// DeactivateResponse represents the soft delete response data.
type DeactivateResponse struct {
	Success bool `json:"success"`
}

// LogoutResponse represents the logout response data.
type LogoutResponse struct {
	Success bool `json:"success"`
}
