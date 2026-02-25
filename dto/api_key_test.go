package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAPIKeyRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateAPIKeyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateAPIKeyRequest{
				Name: "Test Key",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: CreateAPIKeyRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "with expiration",
			req: CreateAPIKeyRequest{
				Name:      "Test Key",
				ExpiresIn: func() *int { i := 30; return &i }(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := ValidateStruct(tt.req)
			if tt.wantErr {
				assert.NotEmpty(t, details)
			} else {
				assert.Empty(t, details)
			}
		})
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	resp := ErrorResponse{
		Error:   "error",
		Message: "test message",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "error")
	assert.Contains(t, string(data), "test message")
}

func TestValidationErrorResponse_JSON(t *testing.T) {
	resp := ValidationErrorResponse{
		Error: "validation error",
		Details: []ValidationErrorDetail{
			{Field: "Name", Message: "This field is required"},
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "validation error")
	assert.Contains(t, string(data), "Name")
}

func TestCreateAPIKeyResponse_JSON(t *testing.T) {
	expiresAt := time.Now().Add(24 * time.Hour)
	resp := CreateAPIKeyResponse{
		ID:        1,
		Key:       "test-key",
		Name:      "Test Key",
		ExpiresAt: &expiresAt,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "test-key")
	assert.Contains(t, string(data), "Test Key")
}

func TestAPIKeyResponse_JSON(t *testing.T) {
	lastUsedAt := time.Now()
	resp := APIKeyResponse{
		ID:         1,
		Name:       "Test Key",
		IsRevoked:  false,
		LastUsedAt: &lastUsedAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Test Key")
}
