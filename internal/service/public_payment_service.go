package service

import (
	"context"

	"github.com/CHainGate/backend/internal/utils"

	"github.com/CHainGate/backend/ethClientApi"
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/google/uuid"
)

type IPublicPaymentService interface {
	HandleNewPayment(priceCurrency enum.FiatCurrency, priceAmount float64, wallet string, mode enum.Mode, callback string, merchant *model.Merchant) (*model.Payment, error)
	HandleNewInvoice(payment *model.Payment, currency enum.CryptoCurrency) (*model.Payment, error)
}

type publicPaymentService struct {
	merchantRepository     repository.IMerchantRepository
	paymentRepository      repository.IPaymentRepository
	internalPaymentService IInternalPaymentService
}

func NewPublicPaymentService(merchantRepository repository.IMerchantRepository, paymentRepository repository.IPaymentRepository, internalPaymentService IInternalPaymentService) IPublicPaymentService {
	return &publicPaymentService{merchantRepository, paymentRepository, internalPaymentService}
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

func (s *publicPaymentService) HandleNewInvoice(initialPayment *model.Payment, currency enum.CryptoCurrency) (*model.Payment, error) {
	m, err := s.merchantRepository.FindById(initialPayment.MerchantId)
	if err != nil {
		return nil, err
	}
	var wallet model.Wallet
	for _, w := range m.Wallets {
		if initialPayment.Mode == w.Mode && w.Currency == currency {
			wallet = w
		}
	}
	initialPayment.Wallet = &wallet
	paymentResponse, err := createEthPayment(initialPayment.PriceCurrency, initialPayment.PriceAmount, initialPayment.Wallet.Address, initialPayment.Mode)
	if err != nil {
		return nil, err
	}

	payment, err := s.handleEthClientResponseUpdate(paymentResponse, initialPayment)
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

	paymentState, ok := enum.ParseStringToStateEnum(*resp.PaymentState)
	if !ok {
		return nil, err
	}
	initialState := model.PaymentState{
		PaymentState: paymentState,
		PayAmount:    model.NewBigIntFromString(resp.PayAmount),
		ActuallyPaid: model.NewBigIntFromInt(0),
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
		Wallet:              &merchant.Wallets[0],
	}
	payment.ID = uuid.New()

	merchant.Payments = append(merchant.Payments, payment)
	err = s.merchantRepository.Update(merchant)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (s *publicPaymentService) handleEthClientResponseUpdate(resp *ethClientApi.PaymentResponse, payment *model.Payment) (*model.Payment, error) {
	blockChainPaymentId, err := uuid.Parse(resp.PaymentId)
	if err != nil {
		return nil, err
	}

	paymentState, ok := enum.ParseStringToStateEnum(*resp.PaymentState)
	if !ok {
		return nil, err
	}
	initialState := model.PaymentState{
		PaymentState: paymentState,
		PayAmount:    model.NewBigIntFromString(resp.PayAmount),
		ActuallyPaid: model.NewBigIntFromInt(0),
	}

	priceCurrency, ok := enum.ParseStringToFiatCurrencyEnum(resp.PriceCurrency)
	if !ok {
		return nil, err
	}
	payCurrency, ok := enum.ParseStringToCryptoCurrencyEnum(resp.PayCurrency)
	if !ok {
		return nil, err
	}
	payment.PriceAmount = resp.PriceAmount
	payment.PriceCurrency = priceCurrency
	payment.PayCurrency = payCurrency
	payment.BlockchainPaymentId = blockChainPaymentId
	payment.PayAddress = resp.PayAddress

	err = s.internalPaymentService.AddNewPaymentState(payment, initialState)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func createEthPayment(priceCurrency enum.FiatCurrency, priceAmount float64, wallet string, mode enum.Mode) (*ethClientApi.PaymentResponse, error) {
	paymentRequest := *ethClientApi.NewPaymentRequest(priceCurrency.String(), priceAmount, wallet, mode.String())
	configuration := ethClientApi.NewConfiguration()
	configuration.Servers[0].URL = utils.Opts.EthereumBaseUrl
	apiClient := ethClientApi.NewAPIClient(configuration)
	resp, _, err := apiClient.PaymentApi.CreatePayment(context.Background()).PaymentRequest(paymentRequest).Execute()
	if err != nil {
		return nil, err
	}
	return resp, nil
}
