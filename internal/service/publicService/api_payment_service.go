/*
 * Public OpenAPI
 *
 * This is the public OpenAPI definition.
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package publicService

import (
	"context"
	"github.com/CHainGate/backend/internal/service"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/CHainGate/backend/publicApi"
	"net/http"
)

// PaymentApiService is a service that implements the logic for the PaymentApiServicer
// This service should implement the business logic for every endpoint for the PaymentApi API.
// Include any external packages or services that will be required by this service.
type PaymentApiService struct {
	authenticationService service.IAuthenticationService
	publicApiService      service.IPublicPaymentService
}

// NewPaymentApiService creates a default api service
func NewPaymentApiService(
	publicApiService service.IPublicPaymentService,
	authenticationService service.IAuthenticationService,
) publicApi.PaymentApiServicer {
	return &PaymentApiService{authenticationService, publicApiService}
}

// NewPayment - Create a new payment
func (s *PaymentApiService) NewPayment(_ context.Context, xAPIKEY string, paymentRequestDto publicApi.PaymentRequestDto) (publicApi.ImplResponse, error) {
	merchant, apiKey, err := s.authenticationService.HandleApiAuthentication(xAPIKEY)
	if err != nil {
		if err.Error() == "not authorized" {
			return publicApi.Response(http.StatusForbidden, nil), err
		}
		return publicApi.Response(http.StatusInternalServerError, nil), err
	}

	priceCurrency, ok := enum.ParseStringToFiatCurrencyEnum(paymentRequestDto.PriceCurrency)
	if !ok {

	}
	payment, err := s.publicApiService.HandleNewPayment(priceCurrency, paymentRequestDto.PriceAmount, "wallet_add", apiKey.Mode, paymentRequestDto.CallbackUrl, merchant)
	if err != nil {
		return publicApi.ImplResponse{}, err
	}

	paymentResponseDto := publicApi.PaymentResponseDto{
		Id:            payment.ID.String(),
		PayAddress:    payment.PayAddress,
		PriceAmount:   payment.PriceAmount,
		PriceCurrency: payment.PriceCurrency.String(),
		PayAmount:     payment.PaymentStates[0].PayAmount,
		PayCurrency:   payment.PayCurrency.String(),
		ActuallyPaid:  &payment.PaymentStates[0].ActuallyPaid,
		CallbackUrl:   payment.CallbackUrl,
		PaymentState:  payment.PaymentStates[0].PaymentState.String(),
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}
	return publicApi.Response(http.StatusCreated, paymentResponseDto), nil
}
