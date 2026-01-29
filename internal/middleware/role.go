package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/response"
)

// RequireRole verifies if the authenticated user has one of the allowed roles
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			if claims == nil {
				response.Unauthorized(w, "Unauthorized: User context missing")
				return
			}

			userRole := claims.Role
			isAllowed := false

			// Check if user has one of the allowed roles
			for _, role := range roles {
				if strings.EqualFold(userRole, role) {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				response.Forbidden(w, fmt.Sprintf("Forbidden: User role '%s' does not have access", userRole))
				return
			}
			
			// Also define context key for domains that need to check it from context? 
			// Not needed for now since we just block if not allowed.

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a shortcut for RequireRole("admin")
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRole(string(domain.RoleAdmin))
}
