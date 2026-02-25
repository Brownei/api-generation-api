package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/controllers"
	"github.com/Brownei/api-generation-api/db"
	"github.com/Brownei/api-generation-api/dto"
	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupControllerTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = database.AutoMigrate(&db.User{})
	require.NoError(t, err)
	return database
}

func TestAuthController_Login_Success(t *testing.T) {
	database := setupControllerTestDB(t)
	cfg := config.LoadAppConfig()
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	userService := services.NewUserService(database, cfg)
	authService := services.NewAuthService(database, cfg)

	hashedPassword, _ := authService.HashPassword("password123")
	user := &db.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	database.Create(user)

	authController := controllers.NewAuthController(userService, authService, sugar)

	body := dto.AuthDto{
		Email:   "test@example.com",
		Pasword: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authController.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_Login_UserNotFound(t *testing.T) {
	database := setupControllerTestDB(t)
	cfg := config.LoadAppConfig()
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	userService := services.NewUserService(database, cfg)
	authService := services.NewAuthService(database, cfg)

	authController := controllers.NewAuthController(userService, authService, sugar)

	body := dto.AuthDto{
		Email:   "nonexistent@example.com",
		Pasword: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authController.Login(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAuthController_Login_InvalidPassword(t *testing.T) {
	database := setupControllerTestDB(t)
	cfg := config.LoadAppConfig()
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	userService := services.NewUserService(database, cfg)
	authService := services.NewAuthService(database, cfg)

	hashedPassword, _ := authService.HashPassword("password123")
	user := &db.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	database.Create(user)

	authController := controllers.NewAuthController(userService, authService, sugar)

	body := dto.AuthDto{
		Email:   "test@example.com",
		Pasword: "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authController.Login(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAuthController_Login_InvalidJSON(t *testing.T) {
	database := setupControllerTestDB(t)
	cfg := config.LoadAppConfig()
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	userService := services.NewUserService(database, cfg)
	authService := services.NewAuthService(database, cfg)

	authController := controllers.NewAuthController(userService, authService, sugar)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authController.Login(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthController_Register_Success(t *testing.T) {
	database := setupControllerTestDB(t)
	cfg := config.LoadAppConfig()
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	userService := services.NewUserService(database, cfg)
	authService := services.NewAuthService(database, cfg)

	authController := controllers.NewAuthController(userService, authService, sugar)

	body := dto.AuthDto{
		Email:   "newuser@example.com",
		Pasword: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authController.Register(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_Register_UserAlreadyExists(t *testing.T) {
	database := setupControllerTestDB(t)
	cfg := config.LoadAppConfig()
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	userService := services.NewUserService(database, cfg)
	authService := services.NewAuthService(database, cfg)

	hashedPassword, _ := authService.HashPassword("password123")
	user := &db.User{
		Name:     "Test User",
		Email:    "existing@example.com",
		Password: hashedPassword,
	}
	database.Create(user)

	authController := controllers.NewAuthController(userService, authService, sugar)

	body := dto.AuthDto{
		Email:   "existing@example.com",
		Pasword: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authController.Register(w, req)

	assert.Equal(t, http.StatusNotModified, w.Code)
}

func TestAPIKeyController_CreateAPIKey_Success(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	database.AutoMigrate(&db.User{}, &db.APIKey{})

	cfg := config.LoadAppConfig()
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	authService := services.NewAuthService(database, cfg)

	hashedPassword, _ := authService.HashPassword("password123")
	user := &db.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	database.Create(user)

	apiKeyService := services.NewAPIKeyService(database)
	apiKeyController := controllers.NewAPIKeyController(apiKeyService, sugar)

	body := dto.CreateAPIKeyRequest{
		Name: "Test Key",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api-key", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), utils.UserIDKey, user.ID))
	w := httptest.NewRecorder()

	apiKeyController.CreateAPIKey(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
