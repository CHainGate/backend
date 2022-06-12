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
	"errors"
	"fmt"
	"github.com/CHainGate/backend/internal/utils"
	"net/http"

	"github.com/CHainGate/backend/internal/service"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/CHainGate/backend/publicApi"
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
		return publicApi.Response(http.StatusBadRequest, nil), errors.New("bad price currency")
	}

	payCurrency, ok := enum.ParseStringToCryptoCurrencyEnum(paymentRequestDto.PayCurrency)
	if !ok {
		return publicApi.Response(http.StatusBadRequest, nil), errors.New("bad pay currency")
	}

	var wallet string
	for _, w := range merchant.Wallets {
		if apiKey.Mode == w.Mode && payCurrency == w.Currency {
			wallet = w.Address
		}
	}

	if wallet == "" {
		errorMessage := fmt.Sprintf("no outcome address defined for %s in mode %s", payCurrency.String(), apiKey.Mode.String())
		return publicApi.Response(http.StatusBadRequest, nil), errors.New(errorMessage)
	}

	payment, err := s.publicApiService.HandleNewPayment(priceCurrency, paymentRequestDto.PriceAmount, payCurrency, wallet, apiKey.Mode, paymentRequestDto.CallbackUrl, merchant)
	if err != nil {
		if err.Error() == "Pay amount is too low " {
			return publicApi.Response(http.StatusBadRequest, nil), err
		}
		return publicApi.Response(http.StatusInternalServerError, nil), err
	}

	payAmount, err := utils.ConvertAmountToBase(payment.PayCurrency, payment.PaymentStates[0].PayAmount.Int)
	if err != nil {
		return publicApi.Response(http.StatusInternalServerError, nil), err
	}
	actuallyPaid, err := utils.ConvertAmountToBase(payment.PayCurrency, payment.PaymentStates[0].ActuallyPaid.Int)
	if err != nil {
		return publicApi.Response(http.StatusInternalServerError, nil), err
	}

	paymentResponseDto := publicApi.PaymentResponseDto{
		Id:            payment.ID.String(),
		PayAddress:    payment.PayAddress,
		PriceAmount:   payment.PriceAmount,
		PriceCurrency: payment.PriceCurrency.String(),
		PayAmount:     payAmount.String(),
		PayCurrency:   payment.PayCurrency.String(),
		ActuallyPaid:  actuallyPaid.String(),
		CallbackUrl:   payment.CallbackUrl,
		PaymentState:  payment.PaymentStates[0].PaymentState.String(),
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}
	return publicApi.Response(http.StatusCreated, paymentResponseDto), nil
}
