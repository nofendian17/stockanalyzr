package middleware

import "context"

// AuthData stores validated authentication data injected by interceptor.
type AuthData struct {
	UserID      string
	AccessToken string
}

type contextKey string

const authDataKey contextKey = "auth_data"

// AuthDataFromContext returns validated auth data from context.
func AuthDataFromContext(ctx context.Context) (AuthData, bool) {
	authData, ok := ctx.Value(authDataKey).(AuthData)
	return authData, ok
}

// UserIDFromContext keeps a focused helper when only user_id is needed.
func UserIDFromContext(ctx context.Context) (string, bool) {
	authData, ok := AuthDataFromContext(ctx)
	if !ok || authData.UserID == "" {
		return "", false
	}
	return authData.UserID, true
}
