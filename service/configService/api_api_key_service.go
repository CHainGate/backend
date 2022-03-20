/*
 * Config OpenAPI
 *
 * This is the config OpenAPI definition.
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package configService

import (
	"CHainGate/backend/configApi"
	"CHainGate/backend/database"
	"CHainGate/backend/models"
	"CHainGate/backend/utils"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

// ApiKeyApiService is a service that implements the logic for the ApiKeyApiServicer
// This service should implement the business logic for every endpoint for the ApiKeyApi API.
// Include any external packages or services that will be required by this service.
type ApiKeyApiService struct {
}

// NewApiKeyApiService creates a default api service
func NewApiKeyApiService() configApi.ApiKeyApiServicer {
	return &ApiKeyApiService{}
}

// DeleteApiKey - delete api key
func (s *ApiKeyApiService) DeleteApiKey(ctx context.Context, apiKeyId string, authorization string) (configApi.ImplResponse, error) {
	user, err := checkAuthorizationAndReturnUser(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	result := database.DB.Model(&models.ApiKey{}).Where("id = ? AND user_id = ?", apiKeyId, user.Id).Update("is_active", false)
	if result.Error != nil {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("")
	}
	return configApi.Response(http.StatusNoContent, nil), nil
}

// GenerateApiKey - create new secret api key
func (s *ApiKeyApiService) GenerateApiKey(ctx context.Context, authorization string, apiKeyRequest configApi.ApiKeyRequest) (configApi.ImplResponse, error) {
	user, err := checkAuthorizationAndReturnUser(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	mode, ok := utils.ParseStringToModeEnum(apiKeyRequest.Mode)
	if !ok {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("mode does not exist")
	}

	apiKeyType, ok := utils.ParseStringToApiKeyTypeEnum(apiKeyRequest.KeyType)
	if !ok {
		return configApi.Response(http.StatusForbidden, nil), errors.New("api key type does not exist")
	}

	key := models.ApiKey{
		Id:        uuid.New(),
		Mode:      mode.String(),
		KeyType:   apiKeyType.String(),
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	// generate random api key
	randomBytes := make([]byte, 64)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), errors.New("Key generation failed ")
	}
	clearTextApiKey := base64.StdEncoding.EncodeToString(randomBytes)

	if apiKeyType == utils.Secret {
		apiKeyBeginning := clearTextApiKey[0:4]
		apiKeyEnding := clearTextApiKey[len(clearTextApiKey)-5:]
		mac := hmac.New(sha512.New, []byte(utils.Opts.ApiKeySecret))
		_, err := io.WriteString(mac, clearTextApiKey)
		if err != nil {
			return configApi.Response(http.StatusInternalServerError, nil), errors.New("Key generation failed ")
		}
		encryptedApiKey := mac.Sum(nil)
		key.EncryptedKey = hex.EncodeToString(encryptedApiKey)
		key.Key = apiKeyBeginning + "..." + apiKeyEnding // show the first and last 4 letters of the secret api key
	} else {
		key.Key = clearTextApiKey
	}

	user.ApiKeys = append(user.ApiKeys, key)
	result := database.DB.Save(&user)
	if result.Error != nil {
		return configApi.Response(http.StatusInternalServerError, nil), errors.New("User could not be updated ")
	}

	apiKeyDto := configApi.ApiKey{
		Id:        key.Id.String(),
		KeyType:   key.KeyType,
		CreatedAt: key.CreatedAt,
	}

	if apiKeyType == utils.Secret {
		apiKeyDto.Key = clearTextApiKey
	} else {
		apiKeyDto.Key = key.Key
	}
	return configApi.Response(http.StatusCreated, apiKeyDto), nil
}

// GetApiKey - gets the api key
func (s *ApiKeyApiService) GetApiKey(ctx context.Context, mode string, keyType string, authorization string) (configApi.ImplResponse, error) {
	user, err := checkAuthorizationAndReturnUser(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	enumMode, ok := utils.ParseStringToModeEnum(mode)
	if !ok {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("mode does not exist")
	}

	enumApiKeyType, ok := utils.ParseStringToApiKeyTypeEnum(keyType)
	if !ok {
		return configApi.Response(http.StatusForbidden, nil), errors.New("api key type does not exist")
	}

	var keys []models.ApiKey
	result := database.DB.Where("user_id = ? and mode = ? and key_type = ?", user.Id, enumMode.String(), enumApiKeyType.String()).Find(&keys)
	if result.Error != nil {
		return configApi.Response(http.StatusInternalServerError, nil), errors.New("")
	}

	var resultList []configApi.ApiKey
	for _, item := range keys {
		resultList = append(resultList, configApi.ApiKey{
			Id:        item.Id.String(),
			Key:       item.Key,
			KeyType:   item.KeyType,
			CreatedAt: item.CreatedAt,
		})
	}

	return configApi.Response(http.StatusOK, resultList), nil
}
