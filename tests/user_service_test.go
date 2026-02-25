package tests

import (
	"testing"

	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/db"
	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserServiceDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = database.AutoMigrate(&db.User{})
	require.NoError(t, err)
	return database
}

func TestUserService_CreateAUser(t *testing.T) {
	database := setupUserServiceDB(t)
	cfg := config.LoadAppConfig()
	service := services.NewUserService(database, cfg)

	user, err := service.CreateAUser("test@example.com", "hashedpassword")

	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestUserService_CreateAUser_Duplicate(t *testing.T) {
	database := setupUserServiceDB(t)
	cfg := config.LoadAppConfig()
	service := services.NewUserService(database, cfg)

	_, err := service.CreateAUser("test@example.com", "hashedpassword")
	require.NoError(t, err)

	_, err = service.CreateAUser("test@example.com", "hashedpassword")

	assert.Error(t, err)
}

func TestUserService_FindThisUser_Success(t *testing.T) {
	database := setupUserServiceDB(t)
	cfg := config.LoadAppConfig()
	service := services.NewUserService(database, cfg)

	user := &db.User{
		Email:    "test@example.com",
		Password: "password123",
	}
	database.Create(user)

	foundUser, err := service.FindThisUser("test@example.com")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", foundUser.Email)
}

func TestUserService_FindThisUser_NotFound(t *testing.T) {
	database := setupUserServiceDB(t)
	cfg := config.LoadAppConfig()
	service := services.NewUserService(database, cfg)

	_, err := service.FindThisUser("nonexistent@example.com")

	assert.ErrorIs(t, err, types.ErrUserNotFound)
}

func TestUserService_FindThisUser_AfterCreate(t *testing.T) {
	database := setupUserServiceDB(t)
	cfg := config.LoadAppConfig()
	service := services.NewUserService(database, cfg)

	_, err := service.CreateAUser("new@example.com", "password123")
	require.NoError(t, err)

	foundUser, err := service.FindThisUser("new@example.com")

	require.NoError(t, err)
	assert.Equal(t, "new@example.com", foundUser.Email)
}
