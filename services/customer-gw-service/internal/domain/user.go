package domain

// User represents a user in the gateway service domain.
// This is a simplified version of the user entity for gateway purposes.
type User struct {
	ID          string
	Email       string
	FullName    string
	PhoneNumber string
	Disabled    bool
	DeletedAt   *int64
	CreatedAt   int64
	UpdatedAt   int64
}

// IsDeleted returns true if user is soft deleted
func (u User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// TokenPair holds access and refresh tokens with their expiry times.
type TokenPair struct {
	AccessToken           string
	AccessTokenExpiresAt  int64
	RefreshToken          string
	RefreshTokenExpiresAt int64
}

// AuthToken represents the authentication token data.
type AuthToken struct {
	AccessToken                 string
	AccessTokenExpiresAtUnixMs  int64
	RefreshToken                string
	RefreshTokenExpiresAtUnixMs int64
}
