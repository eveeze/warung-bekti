package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

type etagResponseWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func (w *etagResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *etagResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

// ETag adds ETag support to GET requests
func ETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Intercept response
		ew := &etagResponseWriter{
			ResponseWriter: w,
			buf:            &bytes.Buffer{},
			statusCode:     http.StatusOK, // default
		}

		next.ServeHTTP(ew, r)

		// Calculate hash
		hash := sha256.Sum256(ew.buf.Bytes())
		etag := `"` + hex.EncodeToString(hash[:]) + `"`

		// Set ETag header
		w.Header().Set("ETag", etag)

		// Check If-None-Match
		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		// Write original status and body
		// Copy headers from captured writer
		// (Already done via embedding, but we need to ensure we don't double write header if we write header explicitly)
		// Actually embedding shares the map reference usually, but safely we should be fine.
		// Wait, WriteHeader wasn't called on original 'w' yet.
		
		if ew.statusCode != 0 {
			w.WriteHeader(ew.statusCode)
		}
		w.Write(ew.buf.Bytes())
	})
}
