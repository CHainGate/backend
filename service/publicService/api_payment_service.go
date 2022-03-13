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
	"CHainGate/backend/publicApi"
	"context"
	"errors"
	"net/http"
)

// PaymentApiService is a service that implements the logic for the PaymentApiServicer
// This service should implement the business logic for every endpoint for the PaymentApi API.
// Include any external packages or services that will be required by this service.
type PaymentApiService struct {
}

// NewPaymentApiService creates a default api service
func NewPaymentApiService() publicApi.PaymentApiServicer {
	return &PaymentApiService{}
}

// NewPayment - Create a new payment
func (s *PaymentApiService) NewPayment(ctx context.Context, paymentRequest publicApi.PaymentRequest) (publicApi.ImplResponse, error) {
	// TODO - update NewPayment with the required logic for this service method.
	// Add api_payment_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(201, Payment{}) or use other options such as http.Ok ...
	//return Response(201, Payment{}), nil

	//TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	//return Response(401, nil),nil

	return publicApi.Response(http.StatusNotImplemented, nil), errors.New("NewPayment method not implemented")
}
