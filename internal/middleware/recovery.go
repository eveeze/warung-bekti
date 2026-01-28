package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/eveeze/warung-backend/internal/pkg/logger"
	"github.com/eveeze/warung-backend/internal/pkg/response"
)

// Recovery middleware recovers from panics and returns 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered: %v\n%s", err, debug.Stack())
				response.InternalServerError(w, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
