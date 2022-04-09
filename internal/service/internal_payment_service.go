package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internalApi"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/CHainGate/backend/proxyClientApi"
)

type IInternalPaymentService interface {
	HandlePaymentUpdate(payment internalApi.PaymentUpdateDto) error
}

type internalPaymentService struct {
	paymentRepository repository.IPaymentRepository
}

func NewInternalPaymentService(paymentRepository repository.IPaymentRepository) IInternalPaymentService {
	return &internalPaymentService{paymentRepository}
}

func (s *internalPaymentService) HandlePaymentUpdate(payment internalApi.PaymentUpdateDto) error {
	currentPayment, err := s.paymentRepository.FindByBlockchainIdAndCurrency(payment.PaymentId, payment.PayCurrency)
	if err != nil {
		return err
	}

	paymentState, ok := enum.ParseStringToStateEnum(payment.PaymentState)
	if !ok {
		return err
	}
	newPaymentState := model.PaymentState{
		PaymentState: paymentState,
		ActuallyPaid: *payment.ActuallyPaid,
		PayAmount:    payment.PayAmount,
	}

	currentPayment.PaymentStates = append(currentPayment.PaymentStates, newPaymentState)

	err = s.paymentRepository.Update(currentPayment)
	if err != nil {
		return err
	}

	currentPayment, err = s.paymentRepository.FindByBlockchainIdAndCurrency(payment.PaymentId, payment.PayCurrency)
	if err != nil {
		return err
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
			PayAmount:     currentState.PayAmount,
			PayCurrency:   payment.PayCurrency.String(),
			ActuallyPaid:  *proxyClientApi.NewNullableFloat64(&currentState.ActuallyPaid),
			PaymentStatus: currentState.PaymentState.String(),
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
