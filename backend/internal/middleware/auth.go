package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const UserIDKey contextKey = "userID"

// Claims defines the custom JWT claims stored in the token.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTCookie creates a signed JWT token and returns a cookie with it.
func NewJWTCookie(userID string, secret []byte) (*http.Cookie, error) {
	expiry := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     "nextup_session",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expiry,
		MaxAge:   int(24 * time.Hour / time.Second),
	}, nil
}

// Auth is an HTTP middleware that validates the JWT session cookie.
// On success it injects the user ID into the request context.
func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("nextup_session")
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext retrieves the user ID injected by the Auth middleware.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}
