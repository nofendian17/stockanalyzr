package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"stockanalyzr/pkg/response"
)

// Ensure context is used
var _ = context.Background()

// AuthUser represents authenticated user data.
type AuthUser struct {
	UserID      string
	Email       string
	AccessToken string
}

const authUserContextKey = "auth_user"

// TokenChecker defines the interface for checking token blacklist status.
type TokenChecker interface {
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// JWTMiddleware creates a Gin middleware for JWT authentication.
// The tokenChecker parameter is optional - if nil, blacklist checking is skipped.
func JWTMiddleware(jwtSecret string, tokenChecker TokenChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Check if token is blacklisted
		if tokenChecker != nil {
			isBlacklisted, err := tokenChecker.IsTokenBlacklisted(c.Request.Context(), tokenString)
			if err != nil {
				response.InternalServerError(c, "failed to validate token")
				c.Abort()
				return
			}
			if isBlacklisted {
				response.Unauthorized(c, "token has been revoked")
				c.Abort()
				return
			}
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "invalid token claims")
			c.Abort()
			return
		}

		userID, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)

		authUser := AuthUser{
			UserID:      userID,
			Email:       email,
			AccessToken: tokenString,
		}

		c.Set(authUserContextKey, authUser)
		c.Next()
	}
}

// GetAuthUser retrieves the authenticated user from the Gin context.
func GetAuthUser(c *gin.Context) (AuthUser, bool) {
	val, ok := c.Get(authUserContextKey)
	if !ok {
		return AuthUser{}, false
	}
	authUser, ok := val.(AuthUser)
	return authUser, ok
}
