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
	"net/http"

	"github.com/CHainGate/backend/internal/utils"

	"github.com/CHainGate/backend/internal/service"

	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/pkg/enum"

	"github.com/CHainGate/backend/configApi"
)

// ApiKeyApiService is a service that implements the logic for the ApiKeyApiServicer
// This service should implement the business logic for every endpoint for the ApiKeyApi API.
// Include any external packages or services that will be required by this service.
type ApiKeyApiService struct {
	authenticationService service.IAuthenticationService
	apiKeyRepository      repository.IApiKeyRepository
	merchantRepository    repository.IMerchantRepository
}

// NewApiKeyApiService creates a default api service
func NewApiKeyApiService(
	authenticationService service.IAuthenticationService,
	apiKeyRepository repository.IApiKeyRepository,
	merchantRepository repository.IMerchantRepository,
) configApi.ApiKeyApiServicer {
	return &ApiKeyApiService{authenticationService, apiKeyRepository, merchantRepository}
}

// DeleteApiKey - delete api key
func (s *ApiKeyApiService) DeleteApiKey(_ context.Context, apiKeyId string, authorization string) (configApi.ImplResponse, error) {
	merchant, err := s.authenticationService.HandleJwtAuthentication(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	err = s.apiKeyRepository.Delete(merchant.ID, apiKeyId)
	if err != nil {
		return configApi.Response(http.StatusBadRequest, nil), err
	}
	return configApi.Response(http.StatusNoContent, nil), nil
}

// GenerateApiKey - create new secret api key
func (s *ApiKeyApiService) GenerateApiKey(_ context.Context, authorization string, apiKeyRequestDto configApi.ApiKeyRequestDto) (configApi.ImplResponse, error) {
	merchant, err := s.authenticationService.HandleJwtAuthentication(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	mode, ok := enum.ParseStringToModeEnum(apiKeyRequestDto.Mode)
	if !ok {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("mode does not exist")
	}

	key, err := s.authenticationService.CreateApiKey(mode)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	merchant.ApiKeys = append(merchant.ApiKeys, *key)
	err = s.merchantRepository.Update(merchant)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), errors.New("Merchant could not be updated ")
	}

	decryptedKey, err := service.Decrypt([]byte(utils.Opts.ApiKeySecret), key.ApiKey)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	apiKeyDto := configApi.ApiKeyResponseDto{
		Id:        key.ID.String(),
		CreatedAt: key.CreatedAt,
		Key:       decryptedKey,
	}

	return configApi.Response(http.StatusCreated, apiKeyDto), nil
}

// GetApiKey - gets the api key
func (s *ApiKeyApiService) GetApiKey(_ context.Context, mode string, authorization string) (configApi.ImplResponse, error) {
	merchant, err := s.authenticationService.HandleJwtAuthentication(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	enumMode, ok := enum.ParseStringToModeEnum(mode)
	if !ok {
		return configApi.Response(http.StatusBadRequest, nil), errors.New("mode does not exist")
	}

	key, err := s.apiKeyRepository.FindByMerchantAndMode(merchant.ID, enumMode)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}
	if key.ApiKey == "" {
		return configApi.Response(http.StatusNoContent, nil), nil
	}

	decryptedKey, err := service.Decrypt([]byte(utils.Opts.ApiKeySecret), key.ApiKey)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	result := configApi.ApiKeyResponseDto{
		Id:        key.ID.String(),
		Key:       decryptedKey,
		CreatedAt: key.CreatedAt,
	}

	return configApi.Response(http.StatusOK, result), nil
}
