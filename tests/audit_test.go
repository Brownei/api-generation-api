package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Brownei/api-generation-api/db"
	"github.com/Brownei/api-generation-api/middleware"
	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuditDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = database.AutoMigrate(&db.User{}, &db.AccessLogs{})
	require.NoError(t, err)
	return database
}

func TestAuditLogService_LogRequest(t *testing.T) {
	database := setupAuditDB(t)

	user := &db.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	database.Create(user)

	auditService := services.NewAuditLogService(database)

	auditService.LogRequest(services.AuditLogEntry{
		UserID:     user.ID,
		Method:     "GET",
		Path:       "/v1/api/users/1",
		StatusCode: 200,
		IPAddress:  "192.168.1.1",
		UserAgent:  "TestAgent",
		Duration:   50,
	})

	time.Sleep(100 * time.Millisecond)

	var log db.AccessLogs
	err := database.Last(&log).Error
	require.NoError(t, err)
	assert.Equal(t, user.ID, log.UserID)
	assert.Equal(t, "GET", log.Method)
	assert.Equal(t, "/v1/api/users/1", log.Path)
	assert.Equal(t, 200, log.StatusCode)
	assert.Equal(t, int64(50), log.Duration)
}

func TestAuditMiddleware_LogsRequest(t *testing.T) {
	database := setupAuditDB(t)

	user := &db.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	database.Create(user)

	auditService := services.NewAuditLogService(database)
	mw := middleware.AuditLogMiddleware(auditService)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/users/1", nil)
	req = req.WithContext(context.WithValue(req.Context(), utils.UserIDKey, user.ID))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(100 * time.Millisecond)

	var log db.AccessLogs
	err := database.Last(&log).Error
	require.NoError(t, err)
	assert.Equal(t, user.ID, log.UserID)
	assert.Equal(t, "GET", log.Method)
	assert.Equal(t, "/v1/api/users/1", log.Path)
	assert.Equal(t, 200, log.StatusCode)
}

func TestAuditMiddleware_SkipsUnauthenticated(t *testing.T) {
	database := setupAuditDB(t)
	auditService := services.NewAuditLogService(database)
	mw := middleware.AuditLogMiddleware(auditService)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/users/1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, nextCalled)

	var count int64
	database.Model(&db.AccessLogs{}).Count(&count)
	assert.Equal(t, int64(0), count)
}
