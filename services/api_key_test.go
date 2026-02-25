package services_test

import (
	"testing"
	"time"

	"github.com/Brownei/api-generation-api/db"
	"github.com/Brownei/api-generation-api/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = database.AutoMigrate(&db.User{}, &db.APIKey{})
	require.NoError(t, err)
	return database
}

func createTestUser(t *testing.T, database *gorm.DB) *db.User {
	user := &db.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	err := database.Create(user).Error
	require.NoError(t, err)
	return user
}

func TestGenerateAPIKey_Success(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", nil)

	require.NoError(t, err)
	assert.NotEmpty(t, apiKey.Key)
	assert.Equal(t, "Test Key", apiKey.Name)
	assert.Equal(t, user.ID, apiKey.UserID)
	assert.False(t, apiKey.IsRevoked)
}

func TestGenerateAPIKey_WithExpiration(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	expiresIn := 30
	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", &expiresIn)

	require.NoError(t, err)
	assert.NotNil(t, apiKey.ExpiresAt)
	assert.True(t, apiKey.ExpiresAt.After(time.Now()))
}

func TestGenerateAPIKey_MaxKeysReached(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	for i := 0; i < 3; i++ {
		_, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
		require.NoError(t, err)
	}

	_, err := service.GenerateAPIKey(user.ID, "Extra Key", nil)
	assert.ErrorIs(t, err, services.ErrTooManyAPIKeys)
}

func TestGenerateAPIKey_KeyGenerationError(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	_, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
	require.NoError(t, err)
}

func TestListAPIKeys(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	_, err := service.GenerateAPIKey(user.ID, "Key 1", nil)
	require.NoError(t, err)
	_, err = service.GenerateAPIKey(user.ID, "Key 2", nil)
	require.NoError(t, err)

	keys, err := service.ListAPIKeys(user.ID)

	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestListAPIKeys_Empty(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	keys, err := service.ListAPIKeys(user.ID)

	require.NoError(t, err)
	assert.Len(t, keys, 0)
}

func TestListAPIKeys_OtherUserKeys(t *testing.T) {
	database := setupTestDB(t)
	user1 := createTestUser(t, database)
	user2 := &db.User{
		Name:     "User 2",
		Email:    "user2@example.com",
		Password: "password123",
	}
	database.Create(user2)
	service := services.NewAPIKeyService(database)

	service.GenerateAPIKey(user1.ID, "Key 1", nil)
	service.GenerateAPIKey(user2.ID, "Key 2", nil)

	keys, err := service.ListAPIKeys(user1.ID)

	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

func TestRevokeAPIKey(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
	require.NoError(t, err)

	err = service.RevokeAPIKey(user.ID, apiKey.ID)

	require.NoError(t, err)
	var revokedKey db.APIKey
	err = database.First(&revokedKey, apiKey.ID).Error
	require.NoError(t, err)
	assert.True(t, revokedKey.IsRevoked)
}

func TestRevokeAPIKey_NotFound(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	err := service.RevokeAPIKey(user.ID, 9999)

	assert.ErrorIs(t, err, services.ErrAPIKeyNotFound)
}

func TestRevokeAPIKey_OtherUserKey(t *testing.T) {
	database := setupTestDB(t)
	user1 := createTestUser(t, database)
	user2 := &db.User{
		Name:     "User 2",
		Email:    "user2@example.com",
		Password: "password123",
	}
	database.Create(user2)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user2.ID, "Key", nil)
	require.NoError(t, err)

	err = service.RevokeAPIKey(user1.ID, apiKey.ID)

	assert.ErrorIs(t, err, services.ErrAPIKeyNotFound)
}

func TestRotateAPIKey(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
	require.NoError(t, err)
	oldKey := apiKey.Key

	newApiKey, err := service.RotateAPIKey(user.ID, apiKey.ID)

	require.NoError(t, err)
	assert.NotEqual(t, oldKey, newApiKey.Key)
	assert.Equal(t, "Test Key", newApiKey.Name)

	var oldKeyResult db.APIKey
	err = database.First(&oldKeyResult, apiKey.ID).Error
	require.NoError(t, err)
	assert.True(t, oldKeyResult.IsRevoked)
}

func TestRotateAPIKey_NotFound(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	_, err := service.RotateAPIKey(user.ID, 9999)

	assert.ErrorIs(t, err, services.ErrAPIKeyNotFound)
}

func TestValidateAPIKey_Valid(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
	require.NoError(t, err)

	validatedKey, err := service.ValidateAPIKey(apiKey.Key)

	require.NoError(t, err)
	assert.Equal(t, apiKey.ID, validatedKey.ID)
}

func TestValidateAPIKey_Invalid(t *testing.T) {
	database := setupTestDB(t)
	service := services.NewAPIKeyService(database)

	_, err := service.ValidateAPIKey("invalid-key")

	assert.ErrorIs(t, err, services.ErrInvalidAPIKey)
}

func TestValidateAPIKey_Revoked(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
	require.NoError(t, err)
	err = service.RevokeAPIKey(user.ID, apiKey.ID)
	require.NoError(t, err)

	_, err = service.ValidateAPIKey(apiKey.Key)

	assert.ErrorIs(t, err, services.ErrAPIKeyRevoked)
}

func TestValidateAPIKey_Expired(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
	require.NoError(t, err)

	expiredTime := time.Now().Add(-24 * time.Hour)
	database.Model(apiKey).Update("expires_at", expiredTime)

	_, err = service.ValidateAPIKey(apiKey.Key)

	assert.ErrorIs(t, err, services.ErrAPIKeyExpired)
}

func TestValidateAPIKey_UpdatesLastUsed(t *testing.T) {
	database := setupTestDB(t)
	user := createTestUser(t, database)
	service := services.NewAPIKeyService(database)

	apiKey, err := service.GenerateAPIKey(user.ID, "Test Key", nil)
	require.NoError(t, err)

	_, err = service.ValidateAPIKey(apiKey.Key)
	require.NoError(t, err)

	var updatedKey db.APIKey
	database.First(&updatedKey, apiKey.ID)
	assert.NotNil(t, updatedKey.LastUsedAt)
}
