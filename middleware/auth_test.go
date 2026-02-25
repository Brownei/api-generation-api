package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/middleware"
	"github.com/Brownei/api-generation-api/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthService(t *testing.T) (*services.AuthService, *config.AppConfig) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.LoadAppConfig()
	return services.NewAuthService(db, cfg), cfg
}

func generateTestToken(t *testing.T, authService *services.AuthService, cfg *config.AppConfig, userID uint, email string, expired bool) string {
	var expiration time.Time
	if expired {
		expiration = time.Now().Add(-1 * time.Hour)
	} else {
		expiration = time.Now().Add(1 * time.Hour)
	}

	claims := services.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return tokenString
}

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	authService, _ := setupAuthService(t)
	mw := middleware.NewAuthMiddleware(authService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	mw.Authenticate(next).ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	authService, _ := setupAuthService(t)
	mw := middleware.NewAuthMiddleware(authService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	mw.Authenticate(next).ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	authService, cfg := setupAuthService(t)
	mw := middleware.NewAuthMiddleware(authService)

	token := generateTestToken(t, authService, cfg, 1, "test@example.com", false)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	nextCalled := false
	var ctxUserID uint
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		ctxUserID = r.Context().Value("user_id").(uint)
	})

	mw.Authenticate(next).ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, uint(1), ctxUserID)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	authService, cfg := setupAuthService(t)
	mw := middleware.NewAuthMiddleware(authService)

	token := generateTestToken(t, authService, cfg, 1, "test@example.com", true)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	mw.Authenticate(next).ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	authService, _ := setupAuthService(t)
	mw := middleware.NewAuthMiddleware(authService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	mw.Authenticate(next).ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
