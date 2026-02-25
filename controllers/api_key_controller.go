package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Brownei/api-generation-api/dto"
	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/utils"
	"github.com/Brownei/api-generation-api/validation"
	"go.uber.org/zap"
)

type APIKeyController struct {
	apiKeyService *services.APIKeyService
	logger        *zap.SugaredLogger
}

func NewAPIKeyController(apiKeyService *services.APIKeyService, logger *zap.SugaredLogger) *APIKeyController {
	return &APIKeyController{
		apiKeyService: apiKeyService,
		logger:        logger,
	}
}

func (h *APIKeyController) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(utils.UserIDKey).(uint)

	var req dto.CreateAPIKeyRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		fmt.Printf("Error: %s", err)
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if details := validation.ValidateStruct(req); len(details) > 0 {
		h.respondWithValidationError(w, details)
		return
	}

	_, err := h.apiKeyService.GetAPIKeyThroughItsName(userID, req.Name)
	if err == nil {
		utils.WriteError(w, http.StatusForbidden, errors.New("Please use another name"))
		return
	}

	apiKey, err := h.apiKeyService.GenerateAPIKey(userID, req.Name, req.ExpiresIn)
	if err != nil {
		if errors.Is(err, services.ErrTooManyAPIKeys) {
			h.respondWithError(w, http.StatusForbidden, "Maximum number of active API keys (3) reached")
			return
		}
		h.logger.Error("Failed to generate API key: ", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to generate API key")
		return
	}

	response := dto.CreateAPIKeyResponse{
		ID:        apiKey.ID,
		Key:       apiKey.Key,
		Name:      apiKey.Name,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: *apiKey.CreatedAt,
	}

	h.respondWithJSON(w, http.StatusCreated, response)
}

func (h *APIKeyController) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userIDContext := r.Context().Value(utils.UserIDKey)
	if userIDContext == nil {
		utils.WriteError(w, 404, errors.New("Unauthorized: no user ID in context"))
		return
	}
	userID := userIDContext.(uint)
	fmt.Printf("userId: %v", userID)

	// First, check and update expired keys
	if err := h.checkAndUpdateExpiredKeys(userID); err != nil {
		h.logger.Error("Failed to check expired keys: ", err)
		// Continue anyway - don't return error to user for this background operation
	}

	// Get all keys (now with updated expiry status)
	keys, err := h.apiKeyService.ListAPIKeys(userID)
	if err != nil {
		h.logger.Error("Failed to list API keys: ", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list API keys")
		return
	}

	var response []dto.APIKeyResponse
	for _, key := range keys {
		response = append(response, dto.APIKeyResponse{
			ID:         key.ID,
			Name:       key.Name,
			IsRevoked:  key.IsRevoked,
			ExpiresAt:  key.ExpiresAt,
			LastUsedAt: key.LastUsedAt,
			CreatedAt:  *key.CreatedAt,
			UpdatedAt:  *key.UpdatedAt,
		})
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// Helper method to check and update expired keys
func (h *APIKeyController) checkAndUpdateExpiredKeys(userID uint) error {
	// Get all keys including expired ones
	allKeys, err := h.apiKeyService.ListAPIKeys(userID) // You might need to add this method
	if err != nil {
		return err
	}

	now := time.Now()
	var expiredKeys []uint

	for _, key := range allKeys {
		// Check if key is expired but not revoked
		if key.ExpiresAt != nil && key.ExpiresAt.Before(now) && !key.IsRevoked {
			expiredKeys = append(expiredKeys, key.ID)
		}
	}

	// Update expired keys if any found
	if len(expiredKeys) > 0 {
		if err := h.apiKeyService.RevokeExpiredKeys(expiredKeys); err != nil {
			return err
		}
		h.logger.Infof("Revoked %d expired keys for user %d", len(expiredKeys), userID)
	}

	return nil
}

func (h *APIKeyController) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(utils.UserIDKey).(uint)

	keyID, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid API key ID")
		return
	}

	err = h.apiKeyService.RevokeAPIKey(userID, uint(keyID))
	if err != nil {
		if errors.Is(err, services.ErrAPIKeyNotFound) {
			h.respondWithError(w, http.StatusNotFound, "API key not found")
			return
		}
		h.logger.Error("Failed to revoke API key: ", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to revoke API key")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "API key revoked successfully"})
}

func (h *APIKeyController) RotateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(utils.UserIDKey).(uint)

	keyID, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid API key ID")
		return
	}

	apiKey, err := h.apiKeyService.RotateAPIKey(userID, uint(keyID))
	if err != nil {
		if errors.Is(err, services.ErrAPIKeyNotFound) {
			h.respondWithError(w, http.StatusNotFound, "API key not found")
			return
		}
		h.logger.Error("Failed to rotate API key: ", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to rotate API key")
		return
	}

	response := dto.CreateAPIKeyResponse{
		ID:        apiKey.ID,
		Key:       apiKey.Key,
		Name:      apiKey.Name,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: *apiKey.CreatedAt,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *APIKeyController) respondWithError(w http.ResponseWriter, code int, message string) {
	utils.WriteJSON(w, code, dto.ErrorResponse{Error: "error", Message: message})
}

func (h *APIKeyController) respondWithValidationError(w http.ResponseWriter, details []validation.ValidationErrorDetail) {
	utils.WriteJSON(w, http.StatusBadRequest, validation.ValidationErrorResponse{
		Error:   "validation error",
		Details: details,
	})
}

func (h *APIKeyController) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}
