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
	"context"
	"errors"
	"github.com/CHainGate/backend/internal/repository/userRepository"
	"net/http"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/models"
	"github.com/CHainGate/backend/internal/utils"
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
func (s *ApiKeyApiService) DeleteApiKey(_ context.Context, apiKeyId string, authorization string) (configApi.ImplResponse, error) {
	user, err := checkAuthorizationAndReturnUser(authorization, userRepository.Repository)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	err = userRepository.Repository.DeleteApiKey(user.Id, apiKeyId)
	if err != nil {
		return configApi.Response(http.StatusBadRequest, nil), err
	}
	return configApi.Response(http.StatusNoContent, nil), nil
}

// GenerateApiKey - create new secret api key
func (s *ApiKeyApiService) GenerateApiKey(_ context.Context, authorization string, apiKeyRequestDto configApi.ApiKeyRequestDto) (configApi.ImplResponse, error) {
	user, err := checkAuthorizationAndReturnUser(authorization, userRepository.Repository)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	mode, ok := utils.ParseStringToModeEnum(apiKeyRequestDto.Mode)
	if !ok {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("mode does not exist")
	}

	apiKeyType, ok := utils.ParseStringToApiKeyTypeEnum(apiKeyRequestDto.KeyType)
	if !ok {
		return configApi.Response(http.StatusForbidden, nil), errors.New("api key type does not exist")
	}

	var key *models.ApiKey

	apiKeyDto := configApi.ApiKeyResponseDto{
		Id:        key.Id.String(),
		KeyType:   key.KeyType,
		CreatedAt: key.CreatedAt,
	}

	apiSecretKey, err := utils.GenerateApiKey()
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	if apiKeyType == utils.Secret {
		key, err = handleSecretApiKey(apiSecretKey, mode, apiKeyType)
		if err != nil {
			return configApi.Response(http.StatusInternalServerError, nil), err
		}
	}

	if apiKeyType == utils.Public {
		key, err = handlePublicApiKey(apiSecretKey, mode, apiKeyType)
		if err != nil {
			return configApi.Response(http.StatusInternalServerError, nil), err
		}
	}

	user.ApiKeys = append(user.ApiKeys, *key)
	err = userRepository.Repository.UpdateUser(user)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), errors.New("User could not be updated ")
	}

	return configApi.Response(http.StatusCreated, apiKeyDto), nil
}

// GetApiKey - gets the api key
func (s *ApiKeyApiService) GetApiKey(_ context.Context, mode string, keyType string, authorization string) (configApi.ImplResponse, error) {
	user, err := checkAuthorizationAndReturnUser(authorization, userRepository.Repository)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	enumMode, ok := utils.ParseStringToModeEnum(mode)
	if !ok {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("mode does not exist")
	}

	enumApiKeyType, ok := utils.ParseStringToApiKeyTypeEnum(keyType)
	if !ok {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("api key type does not exist")
	}

	keys, err := userRepository.Repository.FindApiKeyByUserModeKeyType(user.Id, enumMode, enumApiKeyType)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	var resultList []configApi.ApiKeyResponseDto
	for _, item := range keys {
		resultList = append(resultList, configApi.ApiKeyResponseDto{
			Id:        item.Id.String(),
			Key:       item.ApiKey,
			KeyType:   item.KeyType,
			CreatedAt: item.CreatedAt,
		})
	}

	return configApi.Response(http.StatusOK, resultList), nil
}
