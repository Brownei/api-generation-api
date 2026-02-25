package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/utils"
)

func AuditLogMiddleware(auditService *services.AuditLogService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(utils.UserIDKey)
			if userID == nil {
				next.ServeHTTP(w, r)
				return
			}

			fmt.Printf("user id: %v", userID)

			start := time.Now()
			uw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(uw, r)

			duration := time.Since(start).Milliseconds()

			auditService.LogRequest(services.AuditLogEntry{
				UserID:     userID.(uint),
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: uw.statusCode,
				IPAddress:  r.RemoteAddr,
				UserAgent:  r.UserAgent(),
				Duration:   duration,
			})
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}
