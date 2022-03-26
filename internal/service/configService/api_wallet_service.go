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
	"github.com/CHainGate/backend/configApi"
	"net/http"
)

// WalletApiService is a service that implements the logic for the WalletApiServicer
// This service should implement the business logic for every endpoint for the WalletApi API.
// Include any external packages or services that will be required by this service.
type WalletApiService struct {
}

// NewWalletApiService creates a default api service
func NewWalletApiService() configApi.WalletApiServicer {
	return &WalletApiService{}
}

// AddWallet - add new wallet address
func (s *WalletApiService) AddWallet(ctx context.Context, walletRequestDto configApi.WalletRequestDto) (configApi.ImplResponse, error) {
	// TODO - update AddWallet with the required logic for this service method.
	// Add api_wallet_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(201, Wallet{}) or use other options such as http.Ok ...
	//return Response(201, Wallet{}), nil

	//TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	//return Response(401, nil),nil

	return configApi.Response(http.StatusNotImplemented, nil), errors.New("AddWallet method not implemented")
}

// DeleteWallet - delete wallet
func (s *WalletApiService) DeleteWallet(ctx context.Context, walletId string) (configApi.ImplResponse, error) {
	// TODO - update DeleteWallet with the required logic for this service method.
	// Add api_wallet_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	//return Response(200, nil),nil

	//TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	//return Response(400, nil),nil

	//TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	//return Response(401, nil),nil

	return configApi.Response(http.StatusNotImplemented, nil), errors.New("DeleteWallet method not implemented")
}

// GetWallets - get wallets
func (s *WalletApiService) GetWallets(ctx context.Context, mode string) (configApi.ImplResponse, error) {
	// TODO - update GetWallets with the required logic for this service method.
	// Add api_wallet_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, []Wallet{}) or use other options such as http.Ok ...
	//return Response(200, []Wallet{}), nil

	//TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	//return Response(401, nil),nil

	return configApi.Response(http.StatusNotImplemented, nil), errors.New("GetWallets method not implemented")
}