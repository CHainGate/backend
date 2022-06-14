package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/CHainGate/backend/internal/config"
	"gorm.io/gorm"

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
	apiKeyRepository  repository.IApiKeyRepository
}

func NewInternalPaymentService(paymentRepository repository.IPaymentRepository, apiKeyRepository repository.IApiKeyRepository) IInternalPaymentService {
	return &internalPaymentService{paymentRepository, apiKeyRepository}
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

	err = s.callWebhook(payment)
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

	paymentState, ok := enum.ParseStringToStateEnum(payment.PaymentState)

	if paymentState != enum.PartiallyPaid {
		for _, state := range currentPayment.PaymentStates {
			if state.PaymentState.String() == payment.PaymentState {
				log.Println(fmt.Sprintf("Payment %s with state %s already updated", payment.PaymentId, payment.PaymentState))
				return nil
			}
		}
	}

	if !ok {
		return err
	}
	newPaymentState := model.PaymentState{
		PaymentState: paymentState,
		ActuallyPaid: model.NewBigIntFromString(payment.ActuallyPaid),
		PayAmount:    model.NewBigIntFromString(payment.PayAmount),
	}

	currentPayment.PaymentStates = append(currentPayment.PaymentStates, newPaymentState)

	currentPayment.TxHash = payment.TxHash

	err = s.paymentRepository.Update(currentPayment)
	if err != nil {
		return err
	}

	currentPayment, err = s.paymentRepository.FindByBlockchainIdAndCurrency(payment.PaymentId, payCurrency)
	if err != nil {
		return err
	}

	body := model.SocketBody{
		InitialState:   false,
		Currency:       currentPayment.PayCurrency.String(),
		PayAddress:     currentPayment.PayAddress,
		PayAmount:      currentPayment.PaymentStates[0].PayAmount.String(),
		ActuallyPaid:   currentPayment.PaymentStates[0].ActuallyPaid.String(),
		ExpireTime:     model.GetWaitingCreateDate(currentPayment).Add(15 * time.Minute),
		Mode:           currentPayment.Mode.String(),
		SuccessPageURL: currentPayment.SuccessPageUrl,
		FailurePageURL: currentPayment.FailurePageUrl,
	}

	message := model.Message{MessageType: paymentState.String(), Body: body}
	if pool, ok := config.Pools[currentPayment.ID]; ok {
		pool.Broadcast <- message
	}

	err = s.callWebhook(currentPayment)
	if err != nil {
		return err
	}

	return nil
}

func (s *internalPaymentService) callWebhook(payment *model.Payment) error {
	currentState := payment.PaymentStates[0] //states are sorted
	payAmount, err := utils.ConvertAmountToBaseString(payment.PayCurrency, currentState.PayAmount.Int)
	if err != nil {
		return err
	}
	actuallyPaid, err := utils.ConvertAmountToBaseString(payment.PayCurrency, currentState.ActuallyPaid.Int)
	if err != nil {
		return err
	}
	body := proxyClientApi.WebHookBody{
		Data: proxyClientApi.WebHookData{
			PaymentId:     payment.ID.String(),
			PayAddress:    payment.PayAddress,
			PriceAmount:   payment.PriceAmount,
			PriceCurrency: payment.PriceCurrency.String(),
			PayAmount:     payAmount,
			PayCurrency:   payment.PayCurrency.String(),
			ActuallyPaid:  actuallyPaid,
			PaymentState:  currentState.PaymentState.String(),
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		},
	}

	apiKey, err := s.apiKeyRepository.FindByMerchantAndMode(payment.MerchantId, payment.Mode)
	if err != nil {
		return err
	}

	decryptedKey, err := Decrypt([]byte(utils.Opts.ApiKeySecret), apiKey.ApiKey)

	signature, err := createSignature(body.Data, decryptedKey)
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

func createSignature(data proxyClientApi.WebHookData, secret string) (string, error) {
	mac := hmac.New(sha512.New, []byte(secret))
	jsonData, err := json.Marshal(data)
	_, err = io.WriteString(mac, string(jsonData))
	if err != nil {
		return "", err
	}
	expectedMAC := mac.Sum(nil)
	return hex.EncodeToString(expectedMAC), nil
}
