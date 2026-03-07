package security

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"stockanalyzr/services/user-service/internal/domain"
)

const blacklistPrefix = "jwt_bl:"

// JWTManager generates and validates JWT access and refresh tokens.
type JWTManager struct {
	secretKey              []byte
	issuer                 string
	accessExpiryInMinutes  int
	refreshExpiryInMinutes int
	redisClient            *redis.Client
}

func NewJWTManager(secret, issuer string, accessExpiryMinutes, refreshExpiryMinutes int, rds *redis.Client) *JWTManager {
	return &JWTManager{
		secretKey:              []byte(secret),
		issuer:                 issuer,
		accessExpiryInMinutes:  accessExpiryMinutes,
		refreshExpiryInMinutes: refreshExpiryMinutes,
		redisClient:            rds,
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

func (m *JWTManager) ValidateAccessToken(ctx context.Context, tokenString string) (string, error) {
	// First check if token is blacklisted
	isBlacklisted, err := m.redisClient.Exists(ctx, blacklistPrefix+tokenString).Result()
	if err == nil && isBlacklisted > 0 {
		return "", errors.New("token is revoked")
	}

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

func (m *JWTManager) RevokeToken(ctx context.Context, tokenString string) error {
	// Parse token just to get expiration time so we know how long to keep it in redis
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
	})

	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}

	var ttl time.Duration
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if expFloat, ok := claims["exp"].(float64); ok {
			expTime := time.Unix(int64(expFloat), 0)
			ttl = time.Until(expTime)
		}
	}

	if ttl <= 0 {
		return nil // Already expired, no need to blacklist
	}

	return m.redisClient.Set(ctx, blacklistPrefix+tokenString, "true", ttl).Err()
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
