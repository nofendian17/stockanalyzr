package security

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"stockanalyzr/services/user-service/internal/domain"
)

// JWTManager generates and validates JWT access and refresh tokens.
type JWTManager struct {
	secretKey              []byte
	issuer                 string
	accessExpiryInMinutes  int
	refreshExpiryInMinutes int
}

func NewJWTManager(secret, issuer string, accessExpiryMinutes, refreshExpiryMinutes int) *JWTManager {
	return &JWTManager{
		secretKey:              []byte(secret),
		issuer:                 issuer,
		accessExpiryInMinutes:  accessExpiryMinutes,
		refreshExpiryInMinutes: refreshExpiryMinutes,
	}
}

// CreateTokenPair generates both an access token and a refresh token.
func (m *JWTManager) CreateTokenPair(_ context.Context, userID string) (domain.TokenPair, error) {
	now := time.Now().UTC()

	// Access token
	accessExpiry := now.Add(time.Duration(m.accessExpiryInMinutes) * time.Minute)
	accessToken, err := m.signToken(userID, "access", now, accessExpiry)
	if err != nil {
		return domain.TokenPair{}, err
	}

	// Refresh token
	refreshExpiry := now.Add(time.Duration(m.refreshExpiryInMinutes) * time.Minute)
	refreshToken, err := m.signToken(userID, "refresh", now, refreshExpiry)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiry,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiry,
	}, nil
}

func (m *JWTManager) ValidateAccessToken(_ context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("invalid token subject")
	}

	return sub, nil
}

func (m *JWTManager) signToken(userID, tokenType string, now, expiry time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"iss":  m.issuer,
		"type": tokenType,
		"iat":  now.Unix(),
		"exp":  expiry.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}
