package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/pkg/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging middleware logs HTTP requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		// Create wrapped response writer
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request
		duration := time.Since(start)
		logger.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapped.statusCode,
			"duration":   duration.String(),
			"ip":         r.RemoteAddr,
			"user_agent": r.UserAgent(),
		}).Info("HTTP Request")
	})
}
