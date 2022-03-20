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
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	/*	user, err := checkAuthorizationAndReturnUser(authorization)
		if err != nil {
			return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
		}*/

	return configApi.Response(http.StatusNotImplemented, nil), errors.New("DeleteApiKey method not implemented")
}

// GenerateApiKey - create new secret api key
func (s *ApiKeyApiService) GenerateApiKey(ctx context.Context, authorization string, apiKeyRequest configApi.ApiKeyRequest) (configApi.ImplResponse, error) {
	user, err := checkAuthorizationAndReturnUser(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	//TODO: validate mode and keytype
	key := models.ApiKey{
		Id:        uuid.New(),
		Mode:      apiKeyRequest.Mode,
		KeyType:   apiKeyRequest.KeyType,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	// generate random api key
	randomBytes := make([]byte, 64)
	_, err = rand.Read(randomBytes)
	if err != nil {

	}
	clearTextApiKey := base64.StdEncoding.EncodeToString(randomBytes)

	if apiKeyRequest.KeyType == "secret" {
		apiKeyBeginning := clearTextApiKey[0:4]
		apiKeyEnding := clearTextApiKey[len(clearTextApiKey)-5 : len(clearTextApiKey)-1]
		encryptedApiKey, err := bcrypt.GenerateFromPassword([]byte(clearTextApiKey), bcrypt.DefaultCost)
		if err != nil {
			return configApi.Response(http.StatusInternalServerError, nil), errors.New("Key generation failed ")
		}
		key.EncryptedKey = encryptedApiKey
		key.Key = apiKeyBeginning + "..." + apiKeyEnding // show the first and last 4 letters of the secret api key
	} else {
		key.Key = clearTextApiKey
	}

	user.ApiKey = append(user.ApiKey, key)
	result := database.DB.Save(&user)
	if result.Error != nil {
		return configApi.Response(http.StatusInternalServerError, nil), errors.New("User cound not be updated ")
	}

	apiKeyDto := configApi.ApiKey{
		Id:        key.Id.String(),
		KeyType:   key.KeyType,
		CreatedAt: key.CreatedAt,
	}

	if apiKeyRequest.KeyType == "secret" {
		apiKeyDto.Key = clearTextApiKey
	} else {
		apiKeyDto.Key = key.Key
	}
	return configApi.Response(http.StatusCreated, apiKeyDto), nil
}

// GetApiKey - gets the api key
func (s *ApiKeyApiService) GetApiKey(ctx context.Context, mode string, keyType string, authorization string) (configApi.ImplResponse, error) {
	// TODO - update GetApiKey with the required logic for this service method.
	// Add api_api_key_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, ApiKey{}) or use other options such as http.Ok ...
	//return Response(200, ApiKey{}), nil

	//TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	//return Response(401, nil),nil

	return configApi.Response(http.StatusNotImplemented, nil), errors.New("GetApiKey method not implemented")
}
