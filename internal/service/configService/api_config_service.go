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
	"net/http"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/service"
	"github.com/CHainGate/backend/pkg/enum"
)

// ConfigApiService is a service that implements the logic for the ConfigApiServicer
// This service should implement the business logic for every endpoint for the ConfigApi API.
// Include any external packages or services that will be required by this service.
type ConfigApiService struct {
	authenticationService service.IAuthenticationService
}

// NewConfigApiService creates a default api service
func NewConfigApiService(authenticationService service.IAuthenticationService) configApi.ConfigApiServicer {
	return &ConfigApiService{authenticationService}
}

// GetConfig - Get the configuration
func (s *ConfigApiService) GetConfig(_ context.Context, authorization string) (configApi.ImplResponse, error) {
	currencyDetails := enum.GetCryptoCurrencyDetails()
	supportedCryptoCurrencies := make([]configApi.Currency, 0)
	for _, c := range currencyDetails {
		supportedCryptoCurrencies = append(supportedCryptoCurrencies, configApi.Currency{
			Name:             c.Name,
			ShortName:        c.ShortName,
			ConversionFactor: c.ConversionFactor,
		})
	}

	config := configApi.ConfigResponseDto{
		SupportedCryptoCurrencies: supportedCryptoCurrencies,
	}
	return configApi.Response(http.StatusOK, config), nil
}
