package websocket

import (
	"fmt"
	"net/http"

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

	switch state {
	case enum.CurrencySelection:
		client.SendInitialCoins()
		currency := client.Read()
		payCurrency, _ := enum.ParseStringToCryptoCurrencyEnum(currency)
		publicPaymentService.HandleNewInvoice(payment, payCurrency)
	case enum.Waiting:
		client.SendWaiting(payment)
	case enum.Paid:
		client.SendReceivedTX()
	case enum.Confirmed:
		client.SendConfirmed()
	case enum.Forwarded:
		client.SendConfirmed()
	case enum.Finished:
		client.SendConfirmed()
	case enum.Expired:
		client.SendExpired()
	case enum.Failed:
		client.SendFailed()
	}
	client.Read()
}
