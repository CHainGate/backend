package service

import (
	"context"
	"errors"
	"io/ioutil"
	"time"

	"github.com/CHainGate/backend/internal/config"

	"github.com/CHainGate/backend/internal/utils"

	"github.com/CHainGate/backend/btcClientApi"
	"github.com/CHainGate/backend/ethClientApi"
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/google/uuid"
)

type IPublicPaymentService interface {
	HandleNewPayment(priceCurrency enum.FiatCurrency, priceAmount float64, payCurrency enum.CryptoCurrency, wallet string, mode enum.Mode, callback string, merchant *model.Merchant) (*model.Payment, error)
	HandleNewInvoice(payment *model.Payment, currency enum.CryptoCurrency) (*model.Payment, error)
}

type publicPaymentService struct {
	merchantRepository     repository.IMerchantRepository
	paymentRepository      repository.IPaymentRepository
	internalPaymentService IInternalPaymentService
}

type PaymentResponse struct {
	PaymentId     string
	PaymentState  *string
	PayCurrency   string
	PayAmount     string
	PriceCurrency string
	PriceAmount   float64
	PayAddress    string
}

func NewPublicPaymentService(merchantRepository repository.IMerchantRepository, paymentRepository repository.IPaymentRepository, internalPaymentService IInternalPaymentService) IPublicPaymentService {
	return &publicPaymentService{merchantRepository, paymentRepository, internalPaymentService}
}

func (s *publicPaymentService) HandleNewPayment(priceCurrency enum.FiatCurrency, priceAmount float64, payCurrency enum.CryptoCurrency, wallet string, mode enum.Mode, callback string, merchant *model.Merchant) (*model.Payment, error) {
	var paymentReponse PaymentResponse
	if payCurrency == enum.ETH {
		response, err := createEthPayment(priceCurrency, priceAmount, wallet, mode)
		if err != nil {
			return nil, err
		}

		paymentReponse = PaymentResponse{
			PaymentId:     response.PaymentId,
			PaymentState:  response.PaymentState,
			PayCurrency:   response.PayCurrency,
			PayAmount:     response.PayAmount,
			PriceCurrency: response.PriceCurrency,
			PriceAmount:   response.PriceAmount,
			PayAddress:    response.PayAddress,
		}
	}

	if payCurrency == enum.BTC {
		response, err := createBtcPayment(priceCurrency, priceAmount, wallet, mode)
		if err != nil {
			return nil, err
		}

		paymentReponse = PaymentResponse{
			PaymentId:     response.PaymentId,
			PaymentState:  &response.PaymentState,
			PayCurrency:   response.PayCurrency,
			PayAmount:     response.PayAmount,
			PriceCurrency: response.PriceCurrency,
			PriceAmount:   response.PriceAmount,
			PayAddress:    response.PayAddress,
		}
	}

	payment, err := s.handleBlockchainResponsePayment(&paymentReponse, mode, callback, merchant)
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

	var paymentResponse PaymentResponse
	if currency == enum.ETH {
		response, err := createEthPayment(initialPayment.PriceCurrency, initialPayment.PriceAmount, initialPayment.Wallet.Address, initialPayment.Mode)
		if err != nil {
			return nil, err
		}

		paymentResponse = PaymentResponse{
			PaymentId:     response.PaymentId,
			PaymentState:  response.PaymentState,
			PayCurrency:   response.PayCurrency,
			PayAmount:     response.PayAmount,
			PriceCurrency: response.PriceCurrency,
			PriceAmount:   response.PriceAmount,
			PayAddress:    response.PayAddress,
		}
	}

	if currency == enum.BTC {
		response, err := createBtcPayment(initialPayment.PriceCurrency, initialPayment.PriceAmount, initialPayment.Wallet.Address, initialPayment.Mode)
		if err != nil {
			return nil, err
		}

		paymentResponse = PaymentResponse{
			PaymentId:     response.PaymentId,
			PaymentState:  &response.PaymentState,
			PayCurrency:   response.PayCurrency,
			PayAmount:     response.PayAmount,
			PriceCurrency: response.PriceCurrency,
			PriceAmount:   response.PriceAmount,
			PayAddress:    response.PayAddress,
		}
	}

	payment, err := s.handleBlockchainResponseInvoice(&paymentResponse, initialPayment)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *publicPaymentService) handleBlockchainResponsePayment(resp *PaymentResponse, mode enum.Mode, callbackUrl string, merchant *model.Merchant) (*model.Payment, error) {
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
	payment := model.Payment{
		Base:                model.Base{ID: uuid.New()},
		MerchantId:          merchant.ID,
		Mode:                mode,
		PriceAmount:         resp.PriceAmount,
		PriceCurrency:       priceCurrency,
		PayCurrency:         payCurrency,
		BlockchainPaymentId: blockChainPaymentId,
		PaymentStates:       []model.PaymentState{initialState},
		CallbackUrl:         callbackUrl,
		PayAddress:          resp.PayAddress,
		Wallet:              &merchant.Wallets[0],
	}

	err = s.paymentRepository.Create(&payment)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (s *publicPaymentService) handleBlockchainResponseInvoice(resp *PaymentResponse, payment *model.Payment) (*model.Payment, error) {
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

	body := model.SocketBody{
		InitialState:   false,
		Currency:       payment.PayCurrency.String(),
		PayAddress:     payment.PayAddress,
		PayAmount:      resp.PayAmount,
		ActuallyPaid:   payment.PaymentStates[0].ActuallyPaid.String(),
		ExpireTime:     model.GetWaitingCreateDate(payment).Add(15 * time.Minute),
		Mode:           payment.Mode.String(),
		SuccessPageURL: payment.SuccessPageUrl,
		FailurePageURL: payment.FailurePageUrl,
	}
	message := model.Message{MessageType: paymentState.String(), Body: body}
	if pool, ok := config.Pools[payment.ID]; ok {
		pool.Broadcast <- message
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

func createBtcPayment(priceCurrency enum.FiatCurrency, priceAmount float64, wallet string, mode enum.Mode) (*btcClientApi.PaymentResponseDto, error) {
	paymentRequest := *btcClientApi.NewPaymentRequestDto(priceCurrency.String(), priceAmount, wallet, mode.String())
	configuration := btcClientApi.NewConfiguration()
	configuration.Servers[0].URL = utils.Opts.BitcoinBaseUrl
	apiClient := btcClientApi.NewAPIClient(configuration)
	resp, h, err := apiClient.PaymentApi.CreatePayment(context.Background()).PaymentRequestDto(paymentRequest).Execute()
	if err != nil {
		body, _ := ioutil.ReadAll(h.Body)
		if string(body) == "\"Pay amount is too low \"\n" {
			return nil, errors.New("Pay amount is too low ")
		}
		return nil, err
	}
	return resp, nil
}
