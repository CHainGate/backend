package service

import (
	"context"

	"github.com/CHainGate/backend/ethClientApi"
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/google/uuid"
)

type IPublicPaymentService interface {
	HandleNewPayment(priceCurrency enum.FiatCurrency, priceAmount float64, wallet string, mode enum.Mode, callback string, merchant *model.Merchant) (*model.Payment, error)
}

type publicPaymentService struct {
	merchantRepository repository.IMerchantRepository
}

func NewPublicPaymentService(merchantRepository repository.IMerchantRepository) IPublicPaymentService {
	return &publicPaymentService{merchantRepository}
}

func (s *publicPaymentService) HandleNewPayment(priceCurrency enum.FiatCurrency, priceAmount float64, wallet string, mode enum.Mode, callback string, merchant *model.Merchant) (*model.Payment, error) {
	paymentResponse, err := createEthPayment(priceCurrency, priceAmount, wallet, mode)
	if err != nil {
		return nil, err
	}

	payment, err := s.handleEthClientResponse(paymentResponse, mode, callback, merchant)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *publicPaymentService) handleEthClientResponse(resp *ethClientApi.PaymentResponse, mode enum.Mode, callbackUrl string, merchant *model.Merchant) (*model.Payment, error) {
	blockChainPaymentId, err := uuid.Parse(resp.PaymentId)
	if err != nil {
		return nil, err
	}
	// TODO: get wallet from db
	/*	wallet := models.Wallet{
		Mode: mode.String(),
		Currency: "eth",
		Address: "asdwar88asd",
	}*/

	paymentState, ok := enum.ParseStringToStateEnum(resp.PaymentStatus)
	if !ok {
		return nil, err
	}
	initialState := model.PaymentState{
		PaymentState: paymentState,
		PayAmount:    resp.PayAmount,
		ActuallyPaid: 0,
	}

	priceCurrency, ok := enum.ParseStringToFiatCurrencyEnum(resp.PriceCurrency)
	if !ok {
		return nil, err
	}
	payCurrency, ok := enum.ParseStringToCryptoCurrencyEnum(resp.PayCurrency)
	if !ok {
		return nil, err
	}
	payment := model.Payment{
		Mode:                mode,
		PriceAmount:         resp.PriceAmount,
		PriceCurrency:       priceCurrency,
		PayCurrency:         payCurrency,
		BlockchainPaymentId: blockChainPaymentId,
		PaymentStates:       []model.PaymentState{initialState},
		CallbackUrl:         callbackUrl,
		PayAddress:          resp.PayAddress, //TODO: currently not set from eth service
		Wallet:              merchant.Wallets[0],
	}
	payment.ID = uuid.New()

	merchant.Payments = append(merchant.Payments, payment)
	err = s.merchantRepository.Update(merchant)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func createEthPayment(priceCurrency enum.FiatCurrency, priceAmount float64, wallet string, mode enum.Mode) (*ethClientApi.PaymentResponse, error) {
	paymentRequest := *ethClientApi.NewPaymentRequest(priceCurrency.String(), priceAmount, wallet, mode.String())
	configuration := ethClientApi.NewConfiguration()
	apiClient := ethClientApi.NewAPIClient(configuration)
	resp, _, err := apiClient.PaymentApi.CreatePayment(context.Background()).PaymentRequest(paymentRequest).Execute()
	if err != nil {
		return nil, err
	}
	return resp, nil
}
