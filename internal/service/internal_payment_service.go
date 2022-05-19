package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/CHainGate/backend/internal/config"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log"

	"github.com/CHainGate/backend/internal/utils"

	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internalApi"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/CHainGate/backend/proxyClientApi"
)

type IInternalPaymentService interface {
	HandlePaymentUpdate(payment internalApi.PaymentUpdateDto) error
	AddNewPaymentState(payment *model.Payment, paymentState model.PaymentState) error
}

type internalPaymentService struct {
	paymentRepository repository.IPaymentRepository
}

func NewInternalPaymentService(paymentRepository repository.IPaymentRepository) IInternalPaymentService {
	return &internalPaymentService{paymentRepository}
}

func (s *internalPaymentService) AddNewPaymentState(payment *model.Payment, paymentState model.PaymentState) error {
	payment.PaymentStates = append(payment.PaymentStates, paymentState)

	err := s.paymentRepository.Update(payment)
	if err != nil {
		return err
	}

	payment, err = s.paymentRepository.FindByPaymentId(payment.ID)
	if err != nil {
		return err
	}

	err = callWebhook(payment)
	if err != nil {
		return err
	}

	return nil
}

func (s *internalPaymentService) HandlePaymentUpdate(payment internalApi.PaymentUpdateDto) error {
	payCurrency, ok := enum.ParseStringToCryptoCurrencyEnum(payment.PayCurrency)
	if !ok {

	}
	currentPayment, err := s.paymentRepository.FindByBlockchainIdAndCurrency(payment.PaymentId, payCurrency)
	if err != nil {
		// if the blockchain service creates a new payment but the backend cannot save it to the database
		// we will get an expired update after 15min which is fine and can be ignored, because the buyer
		// never sees the pay address
		if errors.Is(err, gorm.ErrRecordNotFound) && payment.PaymentState == enum.Expired.String() {
			return nil
		}
		return err
	}

	for _, state := range currentPayment.PaymentStates {
		if state.PaymentState.String() == payment.PaymentState {
			log.Println(fmt.Sprintf("Payment %s with state %s already updated", payment.PaymentId, payment.PaymentState))
			return nil
		}
	}

	paymentState, ok := enum.ParseStringToStateEnum(payment.PaymentState)
	if !ok {
		return err
	}
	newPaymentState := model.PaymentState{
		PaymentState: paymentState,
		ActuallyPaid: model.NewBigIntFromString(payment.ActuallyPaid),
		PayAmount:    model.NewBigIntFromString(payment.PayAmount),
	}

	currentPayment.PaymentStates = append(currentPayment.PaymentStates, newPaymentState)

	err = s.paymentRepository.Update(currentPayment)
	if err != nil {
		return err
	}

	currentPayment, err = s.paymentRepository.FindByBlockchainIdAndCurrency(payment.PaymentId, payCurrency)
	if err != nil {
		return err
	}

	message := model.Message{MessageType: paymentState.String(), Body: enum.GetCryptoCurrencyDetails()}
	if pool, ok := config.Pools[currentPayment.ID]; ok {
		pool.Broadcast <- message
	}

	err = callWebhook(currentPayment)
	if err != nil {
		return err
	}

	return nil
}

func callWebhook(payment *model.Payment) error {
	currentState := payment.PaymentStates[0] //states are sorted
	body := proxyClientApi.WebHookBody{
		Data: proxyClientApi.WebHookData{
			PaymentId:     payment.ID.String(),
			PayAddress:    payment.PayAddress,
			PriceAmount:   payment.PriceAmount,
			PriceCurrency: payment.PriceCurrency.String(),
			PayAmount:     currentState.PayAmount.String(),
			PayCurrency:   payment.PayCurrency.String(),
			ActuallyPaid:  currentState.ActuallyPaid.String(),
			PaymentState:  currentState.PaymentState.String(),
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		},
	}

	signature, err := createSignature(body.Data)
	if err != nil {
		return err
	}
	body.Signature = signature

	webhook := *proxyClientApi.NewWebHookRequestDto(payment.CallbackUrl, body)
	configuration := proxyClientApi.NewConfiguration()
	configuration.Servers[0].URL = utils.Opts.ProxyBaseUrl
	apiClient := proxyClientApi.NewAPIClient(configuration)
	_, err = apiClient.WebhookApi.SendWebhook(context.Background()).WebHookRequestDto(webhook).Execute()
	if err != nil {
		return err
	}
	return nil
}

func createSignature(data proxyClientApi.WebHookData) (string, error) {
	//TODO: use merchant secret api key to sign, first we need to save the api key on our side
	mac := hmac.New(sha512.New, []byte("supersecret"))
	jsonData, err := json.Marshal(data)
	_, err = io.WriteString(mac, string(jsonData))
	if err != nil {
		return "", err
	}
	expectedMAC := mac.Sum(nil)
	return hex.EncodeToString(expectedMAC), nil
}
