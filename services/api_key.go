package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Brownei/api-generation-api/db"
	"gorm.io/gorm"
)

const MaxActiveAPIKeys = 3

var (
	ErrTooManyAPIKeys = errors.New("maximum number of active API keys reached")
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrAPIKeyNotOwned = errors.New("API key not owned by user")
	ErrAPIKeyRevoked  = errors.New("API key has been revoked")
	ErrAPIKeyExpired  = errors.New("API key has expired")
	ErrInvalidAPIKey  = errors.New("invalid API key")
)

type APIKeyService struct {
	db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) GenerateAPIKey(userID uint, name string, expiresIn *int) (*db.APIKey, error) {
	var count int64
	s.db.Model(&db.APIKey{}).Where("user_id = ? AND is_revoked = ?", userID, false).Count(&count)
	if count >= MaxActiveAPIKeys {
		return nil, ErrTooManyAPIKeys
	}

	key, err := generateRandomKey()
	if err != nil {
		return nil, err
	}

	var expiresAt *time.Time
	if expiresIn != nil && *expiresIn > 0 {
		t := time.Now().Add(time.Duration(*expiresIn) * 24 * time.Hour)
		expiresAt = &t
	}

	apiKey := &db.APIKey{
		Key:       key,
		UserID:    userID,
		Name:      name,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(apiKey).Error; err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *APIKeyService) ListAPIKeys(userID uint) ([]db.APIKey, error) {
	var keys []db.APIKey
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *APIKeyService) RevokeAPIKey(userID, keyID uint) error {
	result := s.db.Model(&db.APIKey{}).Where("id = ? AND user_id = ?", keyID, userID).Update("is_revoked", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (s *APIKeyService) RotateAPIKey(userID, keyID uint) (*db.APIKey, error) {
	var apiKey db.APIKey
	if err := s.db.Where("id = ? AND user_id = ?", keyID, userID).First(&apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	s.db.Model(&db.APIKey{}).Where("id = ?", keyID).Update("is_revoked", true)

	return s.GenerateAPIKey(userID, apiKey.Name, nil)
}

func (s *APIKeyService) ValidateAPIKey(key string) (*db.APIKey, error) {
	var apiKey db.APIKey
	if err := s.db.Where("key = ?", key).First(&apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidAPIKey
		}
		return nil, err
	}

	if apiKey.IsRevoked {
		return nil, ErrAPIKeyRevoked
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}

	s.db.Model(&db.APIKey{}).Where("id = ?", apiKey.ID).Update("last_used_at", time.Now())

	return &apiKey, nil
}

func generateRandomKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *APIKeyService) GetAPIKeyThroughItsName(userID uint, name string) (*db.APIKey, error) {
	var apiKey db.APIKey

	if err := s.db.Where("name = ? AND user_id = ?", name, userID).First(&apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	return &apiKey, nil
}

func (s *APIKeyService) RevokeExpiredKeys(keyIDs []uint) error {
	if len(keyIDs) == 0 {
		return nil
	}

	// Update expired keys to revoked
	result := s.db.Model(&db.APIKey{}).
		Where("id IN ?", keyIDs).
		Update("is_revoked", true)

	return result.Error
}
