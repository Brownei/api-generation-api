package dto

import (
	"time"
)

type CreateAPIKeyRequest struct {
	Name      string `json:"name" validate:"required,max=100"`
	ExpiresIn *int   `json:"expires_in" validate:"omitempty,min=1"`
}

type CreateAPIKeyResponse struct {
	ID        uint       `json:"id"`
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type APIKeyResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	IsRevoked  bool       `json:"is_revoked"`
	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type RevokeAPIKeyRequest struct {
	KeyID uint `json:"key_id" validate:"required,min=1"`
}

type RotateAPIKeyRequest struct {
	KeyID uint `json:"key_id" validate:"required,min=1"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
