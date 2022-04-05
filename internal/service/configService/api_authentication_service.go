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
	"github.com/CHainGate/backend/internal/service"
	"net/http"

	"github.com/CHainGate/backend/configApi"
)

// AuthenticationApiService is a service that implements the logic for the AuthenticationApiServicer
// This service should implement the business logic for every endpoint for the AuthenticationApi API.
// Include any external packages or services that will be required by this service.
type AuthenticationApiService struct {
	authenticationService service.IAuthenticationService
}

// NewAuthenticationApiService creates a default api service
func NewAuthenticationApiService(authenticationService service.IAuthenticationService) configApi.AuthenticationApiServicer {
	return &AuthenticationApiService{authenticationService}
}

// Login - Authenticate to chaingate
func (s *AuthenticationApiService) Login(_ context.Context, loginRequestDto configApi.LoginRequestDto) (configApi.ImplResponse, error) {
	token, err := s.authenticationService.HandleLogin(loginRequestDto.Email, loginRequestDto.Password)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	tokenDto := configApi.TokenResponseDto{Token: token}

	return configApi.Response(http.StatusCreated, tokenDto), nil
}

// Logout - Logs out the merchant
func (s *AuthenticationApiService) Logout(_ context.Context) (configApi.ImplResponse, error) {
	return configApi.Response(http.StatusNotImplemented, nil), errors.New("Logout method not implemented")
}

// RegisterMerchant - Merchant registration
func (s *AuthenticationApiService) RegisterMerchant(_ context.Context, registerRequestDto configApi.RegisterRequestDto) (configApi.ImplResponse, error) {
	err := s.authenticationService.CreateMerchant(registerRequestDto)
	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"merchants_email_key\" (SQLSTATE 23505)" {
			return configApi.Response(http.StatusBadRequest, nil), errors.New("E-Mail already exists")
		}
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	return configApi.Response(http.StatusNoContent, nil), nil
}

// VerifyEmail - Verify merchant E-Mail
func (s *AuthenticationApiService) VerifyEmail(_ context.Context, email string, verificationCode int64) (configApi.ImplResponse, error) {
	err := s.authenticationService.HandleVerification(email, verificationCode)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	return configApi.Response(http.StatusOK, nil), nil
}
