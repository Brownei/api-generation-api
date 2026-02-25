package services_test

import (
	"testing"
	"time"

	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/db"
	"github.com/Brownei/api-generation-api/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthServiceDB(t *testing.T) (*gorm.DB, *config.AppConfig) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = database.AutoMigrate(&db.User{})
	require.NoError(t, err)
	cfg := config.LoadAppConfig()
	return database, cfg
}

func TestAuthService_GenerateToken(t *testing.T) {
	database, cfg := setupAuthServiceDB(t)
	service := services.NewAuthService(database, cfg)

	token, err := service.GenerateToken(1, "test@example.com")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_ValidateToken_Valid(t *testing.T) {
	database, cfg := setupAuthServiceDB(t)
	service := services.NewAuthService(database, cfg)

	token, err := service.GenerateToken(1, "test@example.com")
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)

	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	database, cfg := setupAuthServiceDB(t)
	service := services.NewAuthService(database, cfg)

	_, err := service.ValidateToken("invalid.token.here")

	assert.Error(t, err)
}

func TestAuthService_GetUserByID(t *testing.T) {
	database, cfg := setupAuthServiceDB(t)
	service := services.NewAuthService(database, cfg)

	user := &db.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	database.Create(user)

	foundUser, err := service.GetUserByID(user.ID)

	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, "test@example.com", foundUser.Email)
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	database, cfg := setupAuthServiceDB(t)
	service := services.NewAuthService(database, cfg)

	_, err := service.GetUserByID(999)

	assert.ErrorIs(t, err, services.ErrUserNotFound)
}

func TestAuthService_HashPassword(t *testing.T) {
	database, cfg := setupAuthServiceDB(t)
	service := services.NewAuthService(database, cfg)

	hash, err := service.HashPassword("password123")

	require.NoError(t, err)
	assert.NotEqual(t, "password123", hash)
	assert.NotEmpty(t, hash)
}

func TestAuthService_CheckPassword(t *testing.T) {
	database, cfg := setupAuthServiceDB(t)
	service := services.NewAuthService(database, cfg)

	hash, _ := service.HashPassword("password123")

	assert.True(t, service.CheckPassword("password123", hash))
	assert.False(t, service.CheckPassword("wrongpassword", hash))
}

func TestClaims_Expired(t *testing.T) {
	claims := services.Claims{
		UserID: 1,
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	cfg := config.LoadAppConfig()
	database, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	service := services.NewAuthService(database, cfg)

	_, err := service.ValidateToken(tokenString)

	assert.Error(t, err)
}
