package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/middleware"
)

func generateTestToken(userID, username, role, secret string, exp time.Duration) string {
	claims := &domain.UserClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestAuthMiddleware(t *testing.T) {
	jwtCfg := &config.JWTConfig{Secret: "test-secret-key-12345"}
	uid := uuid.New().String()
	token := generateTestToken(uid, "testuser", "admin", jwtCfg.Secret, time.Hour)

	handler := middleware.Auth(jwtCfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUserFromContext(r.Context())
		if user == nil {
			t.Errorf("expected user in context")
		}
		extractedUID, ok := middleware.GetUserID(r.Context())
		if !ok || extractedUID.String() != uid {
			t.Errorf("expected extracted user ID to match")
		}
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Success with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// 2. Missing header
	reqNoAuth := httptest.NewRequest("GET", "/test", nil)
	recNoAuth := httptest.NewRecorder()
	handler.ServeHTTP(recNoAuth, reqNoAuth)
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recNoAuth.Code)
	}

	// 3. Bad header format
	reqBadAuth := httptest.NewRequest("GET", "/test", nil)
	reqBadAuth.Header.Set("Authorization", "Token "+token)
	recBadAuth := httptest.NewRecorder()
	handler.ServeHTTP(recBadAuth, reqBadAuth)
	if recBadAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recBadAuth.Code)
	}

	// 4. Invalid token
	reqInvalid := httptest.NewRequest("GET", "/test", nil)
	reqInvalid.Header.Set("Authorization", "Bearer invalid.token.value")
	recInvalid := httptest.NewRecorder()
	handler.ServeHTTP(recInvalid, reqInvalid)
	if recInvalid.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recInvalid.Code)
	}

	// 5. OptionalAuth
	optHandler := middleware.OptionalAuth(jwtCfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	reqOpt := httptest.NewRequest("GET", "/test", nil)
	recOpt := httptest.NewRecorder()
	optHandler.ServeHTTP(recOpt, reqOpt)
	if recOpt.Code != http.StatusOK {
		t.Errorf("expected status 200 for optional auth without header")
	}

	reqOptValid := httptest.NewRequest("GET", "/test", nil)
	reqOptValid.Header.Set("Authorization", "Bearer "+token)
	recOptValid := httptest.NewRecorder()
	optHandler.ServeHTTP(recOptValid, reqOptValid)
	if recOptValid.Code != http.StatusOK {
		t.Errorf("expected status 200 for optional auth with header")
	}

	// Helper when context is empty
	if middleware.GetUserFromContext(context.Background()) != nil {
		t.Errorf("expected nil for empty context")
	}
	if _, ok := middleware.GetUserID(context.Background()); ok {
		t.Errorf("expected false for empty context")
	}
}

func TestRoleMiddleware(t *testing.T) {
	jwtCfg := &config.JWTConfig{Secret: "test-secret-key-12345"}
	adminToken := generateTestToken(uuid.New().String(), "admin", "admin", jwtCfg.Secret, time.Hour)
	cashierToken := generateTestToken(uuid.New().String(), "cashier", "cashier", jwtCfg.Secret, time.Hour)

	adminOnly := middleware.Auth(jwtCfg)(middleware.RequireAdmin()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	roleCheck := middleware.Auth(jwtCfg)(middleware.RequireRole("cashier", "inventory")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Admin accesses adminOnly
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	adminOnly.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Cashier accesses adminOnly -> 403
	reqCashier := httptest.NewRequest("GET", "/admin", nil)
	reqCashier.Header.Set("Authorization", "Bearer "+cashierToken)
	recCashier := httptest.NewRecorder()
	adminOnly.ServeHTTP(recCashier, reqCashier)
	if recCashier.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", recCashier.Code)
	}

	// Cashier accesses RequireRole cashier
	recRole := httptest.NewRecorder()
	roleCheck.ServeHTTP(recRole, reqCashier)
	if recRole.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recRole.Code)
	}

	// Admin accesses RequireRole cashier -> 403
	recRoleAdmin := httptest.NewRecorder()
	roleCheck.ServeHTTP(recRoleAdmin, req)
	if recRoleAdmin.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", recRoleAdmin.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// OPTIONS preflight
	reqOptions := httptest.NewRequest("OPTIONS", "/test", nil)
	recOptions := httptest.NewRecorder()
	handler.ServeHTTP(recOptions, reqOptions)
	if recOptions.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", recOptions.Code)
	}
	if recOptions.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("missing CORS header")
	}

	// Normal request
	reqGet := httptest.NewRequest("GET", "/test", nil)
	recGet := httptest.NewRecorder()
	handler.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Errorf("expected 200 for GET, got %d", recGet.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went catastrophically wrong!")
	}))

	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestETagMiddleware(t *testing.T) {
	handler := middleware.ETag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message": "hello etag"}`))
	}))

	// Initial request
	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header")
	}

	// If-None-Match with matching ETag -> 304
	req304 := httptest.NewRequest("GET", "/api/v1/products", nil)
	req304.Header.Set("If-None-Match", etag)
	rec304 := httptest.NewRecorder()
	handler.ServeHTTP(rec304, req304)
	if rec304.Code != http.StatusNotModified {
		t.Errorf("expected 304, got %d", rec304.Code)
	}

	// Test Flush on etag writer
	flushHandler := middleware.ETag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		w.Write([]byte("flushed"))
	}))
	reqFlush := httptest.NewRequest("GET", "/api/v1/flushed", nil)
	recFlush := httptest.NewRecorder()
	flushHandler.ServeHTTP(recFlush, reqFlush)

	// Streaming bypass (/api/v1/events)
	eventHandler := middleware.ETag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		w.Write([]byte("event: ping\n\n"))
	}))
	reqSSE := httptest.NewRequest("GET", "/api/v1/events", nil)
	recSSE := httptest.NewRecorder()
	eventHandler.ServeHTTP(recSSE, reqSSE)
	if recSSE.Code != http.StatusOK {
		t.Errorf("expected 200 for SSE, got %d", recSSE.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	handler := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("logged"))
	}))

	req := httptest.NewRequest("GET", "/test-log", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID header")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// Limit of 2 requests per 20ms
	handler := middleware.RateLimit(2, 20*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("expected 200 for request 1")
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 for request 2")
	}

	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for request 3, got %d", rec3.Code)
	}

	// Sleep beyond window so rate limiter cleans up and allows again
	time.Sleep(30 * time.Millisecond)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req1)
	if rec4.Code != http.StatusOK {
		t.Errorf("expected 200 after window reset")
	}
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for request 3, got %d", rec3.Code)
	}
}

func TestAuditMiddleware(t *testing.T) {
	handler := middleware.Audit(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/audit", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID")
	}

	// Test LogAudit with empty context
	middleware.LogAudit(context.Background(), nil, "TEST", "product", nil, "notes")
}
