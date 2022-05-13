package websocket

import (
	"fmt"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internal/service"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/google/uuid"
	"net/http"
)

func ServeWs(pool *Pool, w http.ResponseWriter, r *http.Request, publicPaymentService service.IPublicPaymentService, paymentRepository repository.IPaymentRepository, paymentId uuid.UUID) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &Client{
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
		client.SendWaiting()
	case enum.Waiting:
		client.SendWaiting()
	case enum.Paid:
		client.SendReceivedTX()
	case enum.Confirmed:
		client.SendConfirmed()
	}
}
