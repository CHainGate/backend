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

	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internal/service"
	"github.com/CHainGate/backend/pkg/enum"

	"github.com/CHainGate/backend/configApi"
)

// LoggingApiService is a service that implements the logic for the LoggingApiServicer
// This service should implement the business logic for every endpoint for the LoggingApi API.
// Include any external packages or services that will be required by this service.
type LoggingApiService struct {
	authenticationService service.IAuthenticationService
	paymentRepository     repository.IPaymentRepository
}

// NewLoggingApiService creates a default api service
func NewLoggingApiService(
	authenticationService service.IAuthenticationService,
	paymentRepository repository.IPaymentRepository,
) configApi.LoggingApiServicer {
	return &LoggingApiService{authenticationService, paymentRepository}
}

// GetLoggingInformation - get logging information
func (s *LoggingApiService) GetLoggingInformation(_ context.Context, mode string, authorization string) (configApi.ImplResponse, error) {
	merchant, err := s.authenticationService.HandleJwtAuthentication(authorization)
	if err != nil {
		return configApi.Response(http.StatusForbidden, nil), errors.New("not authorized")
	}

	parsedMode, ok := enum.ParseStringToModeEnum(mode)
	if !ok {
		return configApi.Response(http.StatusInternalServerError, nil), errors.New("wrong mode")
	}
	payments, err := s.paymentRepository.FindByMerchantIdAndMode(merchant.ID, parsedMode)
	if err != nil {
		return configApi.Response(http.StatusInternalServerError, nil), err
	}

	var result []configApi.LoggingResponseDto
	for _, payment := range payments {
		var history []configApi.PaymentHistory
		for _, state := range payment.PaymentStates {
			actuallyPaid := state.ActuallyPaid
			h := configApi.PaymentHistory{
				Id:            state.ID.String(),
				CreatedAt:     state.CreatedAt,
				PayCurrency:   payment.PayCurrency.String(),
				PaymentState:  state.PaymentState.String(),
				PayAmount:     state.PayAmount.String(),
				ActuallyPaid:  actuallyPaid.String(),
				PriceCurrency: payment.PriceCurrency.String(),
				PriceAmount:   payment.PriceAmount,
				PayAddress:    payment.PayAddress,
			}
			history = append(history, h)
		}

		p := configApi.LoggingResponseDto{
			UpdatedAt:   payment.UpdatedAt,
			CreatedAt:   payment.CreatedAt,
			PaymentId:   payment.ID.String(),
			CallbackUrl: payment.CallbackUrl,
			Transaction: payment.TxHash,
			History:     history,
		}
		result = append(result, p)
	}

	return configApi.Response(http.StatusOK, result), nil
}
