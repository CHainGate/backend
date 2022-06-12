package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CHainGate/backend/internal/model"

	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internal/service"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/google/uuid"
)

func ServeWs(pool *model.Pool, w http.ResponseWriter, r *http.Request, publicPaymentService service.IPublicPaymentService, paymentRepository repository.IPaymentRepository, paymentId uuid.UUID) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &model.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client

	payment, err := paymentRepository.FindByPaymentId(paymentId)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	state := payment.PaymentStates[0].PaymentState
	var body model.SocketBody
	if state != enum.CurrencySelection {
		body = model.SocketBody{
			InitialState:   true,
			Currency:       payment.PayCurrency.String(),
			PayAddress:     payment.PayAddress,
			PayAmount:      payment.PaymentStates[0].PayAmount.String(),
			ActuallyPaid:   payment.PaymentStates[0].ActuallyPaid.String(),
			ExpireTime:     model.GetWaitingCreateDate(payment).Add(15 * time.Minute),
			Mode:           payment.Mode.String(),
			SuccessPageURL: payment.SuccessPageUrl,
			FailurePageURL: payment.FailurePageUrl,
		}
	}

	switch state {
	case enum.CurrencySelection:
		client.SendInitialCoins()
		currency := client.Read()
		payCurrency, _ := enum.ParseStringToCryptoCurrencyEnum(currency)
		publicPaymentService.HandleNewInvoice(payment, payCurrency)
	case enum.Waiting:
		client.SendWaiting(body)
	case enum.PartiallyPaid:
		client.SendPartiallyPaid(body)
	case enum.Paid:
		client.SendReceivedTX(body)
	case enum.Confirmed:
		client.SendConfirmed(body)
	case enum.Forwarded:
		client.SendConfirmed(body)
	case enum.Finished:
		client.SendConfirmed(body)
	case enum.Expired:
		client.SendExpired(body)
	case enum.Failed:
		client.SendFailed(body)
	}
	client.Read()
}
