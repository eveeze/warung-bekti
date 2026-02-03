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
		// Intercept response for all methods to generate ETag
		ew := &etagResponseWriter{
			ResponseWriter: w,
			buf:            &bytes.Buffer{},
			statusCode:     http.StatusOK, // default
		}

		next.ServeHTTP(ew, r)

		// Skip ETag for empty bodies or specific status codes if needed
		// For now, we generate it for everything successful that has content
		if ew.buf.Len() == 0 {
			if ew.statusCode != 0 {
				w.WriteHeader(ew.statusCode)
			}
			return
		}

		// Calculate hash
		hash := sha256.Sum256(ew.buf.Bytes())
		etag := `"` + hex.EncodeToString(hash[:]) + `"`

		// Set ETag header
		w.Header().Set("ETag", etag)

		// Check If-None-Match (only for GET/HEAD mainly, but standard says 304 is for GET/HEAD)
		// For writes, we generally don't return 304, but we DO want to return the ETag.
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			if match := r.Header.Get("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}

		// Write original status and body
		if ew.statusCode != 0 {
			w.WriteHeader(ew.statusCode)
		}
		w.Write(ew.buf.Bytes())
	})
}
