package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/types"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		claims, err := m.authService.ValidateToken(parts[1])
		if err != nil {
			if errors.Is(err, types.ErrTokenExpired) {
				http.Error(w, `{"error": "token expired"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
