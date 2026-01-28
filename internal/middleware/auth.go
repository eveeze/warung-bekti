package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/pkg/response"
)

type contextKey string

const UserContextKey contextKey = "user"

// UserClaims represents JWT claims
type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Auth middleware validates JWT tokens
func Auth(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Unauthorized(w, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.Secret), nil
			})

			if err != nil || !token.Valid {
				response.Unauthorized(w, "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(*UserClaims)
			if !ok {
				response.Unauthorized(w, "Invalid token claims")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware validates JWT if present, but doesn't require it
func OptionalAuth(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				next.ServeHTTP(w, r)
				return
			}

			token, err := jwt.ParseWithClaims(parts[1], &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.Secret), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*UserClaims); ok {
					ctx := context.WithValue(r.Context(), UserContextKey, claims)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves user claims from context
func GetUserFromContext(ctx context.Context) *UserClaims {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}
