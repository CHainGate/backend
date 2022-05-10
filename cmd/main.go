package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/CHainGate/backend/pkg/enum"

	"github.com/google/uuid"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internal/service"
	"github.com/CHainGate/backend/internal/service/configService"
	"github.com/CHainGate/backend/internal/service/internalService"
	"github.com/CHainGate/backend/internal/service/publicService"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/CHainGate/backend/internalApi"
	"github.com/CHainGate/backend/publicApi"
	"github.com/gorilla/websocket"

	"github.com/rs/cors"
)

type SocketMessage struct {
	MessageType string      `json:"type"`
	Data        interface{} `json:"data"`
}

type CurrencySelection struct {
	Currency string `json:"currency"`
}

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var lock = sync.Mutex{}
var clients = make(map[uuid.UUID][]*websocket.Conn)
var clientsCopy = make(map[uuid.UUID][]*websocket.Conn)

func main() {
	utils.NewOpts() // create utils.Opts (env variables)
	merchantRepo, apiKeyRepo, paymentRepo, err := repository.SetupDatabase()
	if err != nil {
		panic(err)
	}

	authService := service.NewAuthenticationService(merchantRepo, apiKeyRepo)
	// config api
	ApiKeyApiService := configService.NewApiKeyApiService(authService, apiKeyRepo, merchantRepo)
	ApiKeyApiController := configApi.NewApiKeyApiController(ApiKeyApiService)

	AuthenticationApiService := configService.NewAuthenticationApiService(authService)
	AuthenticationApiController := configApi.NewAuthenticationApiController(AuthenticationApiService)

	LoggingApiService := configService.NewLoggingApiService(authService, paymentRepo)
	LoggingApiController := configApi.NewLoggingApiController(LoggingApiService)

	WalletApiService := configService.NewWalletApiService(authService, merchantRepo)
	WalletApiController := configApi.NewWalletApiController(WalletApiService)

	ConfigApiService := configService.NewConfigApiService(authService)
	ConfigApiController := configApi.NewConfigApiController(ConfigApiService)

	configRouter := configApi.NewRouter(ApiKeyApiController, AuthenticationApiController, LoggingApiController, WalletApiController, ConfigApiController)

	// public api
	publicPaymentService := service.NewPublicPaymentService(merchantRepo)
	publicInvoiceService := publicService.NewInvoiceApiService(publicPaymentService, authService, merchantRepo)
	PaymentApiService := publicService.NewPaymentApiService(publicPaymentService, authService)
	PaymentApiController := publicApi.NewPaymentApiController(PaymentApiService)
	InvoiceApiController := publicApi.NewInvoiceApiController(publicInvoiceService)

	publicRouter := publicApi.NewRouter(PaymentApiController, InvoiceApiController)

	// internal api
	internalPaymentService := service.NewInternalPaymentService(paymentRepo)
	PaymentUpdateApiService := internalService.NewPaymentUpdateApiService(internalPaymentService)
	PaymentUpdateApiController := internalApi.NewPaymentUpdateApiController(PaymentUpdateApiService)

	internalRouter := internalApi.NewRouter(PaymentUpdateApiController)

	http.Handle("/api/config/", cors.AllowAll().Handler(configRouter))
	http.Handle("/api/public/", cors.AllowAll().Handler(publicRouter))
	http.Handle("/api/internal/", cors.AllowAll().Handler(internalRouter))

	// https://ribice.medium.com/serve-swaggerui-within-your-golang-application-5486748a5ed4
	configFs := http.FileServer(http.Dir("./swaggerui/config"))
	http.Handle("/api/config/swaggerui/", http.StripPrefix("/api/config/swaggerui/", configFs))
	publicFs := http.FileServer(http.Dir("./swaggerui/public"))
	http.Handle("/api/public/swaggerui/", http.StripPrefix("/api/public/swaggerui/", publicFs))
	internalFs := http.FileServer(http.Dir("./swaggerui/internal"))
	http.Handle("/api/internal/swaggerui/", http.StripPrefix("/api/internal/swaggerui/", internalFs))
	http.HandleFunc("/ws", wsEndpoint)

	log.Println("Starting backend-service on port " + strconv.Itoa(utils.Opts.ServerPort))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(utils.Opts.ServerPort), nil))
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("could not upgrade connection: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	paymentIdParam := r.URL.Query().Get("pid")
	paymentId := uuid.MustParse(paymentIdParam)

	lock.Lock()
	clients[paymentId] = append(clients[paymentId], conn)
	lock.Unlock()
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("closing connection for %v", paymentId)
		lock.Lock()
		delete(clients, paymentId)
		lock.Unlock()
		return nil
	})

	conn.SetPongHandler(func(appData string) error {
		log.Printf(appData)
		return nil
	})

	go notifyBrowser(paymentId)
}

func notifyBrowser(pid uuid.UUID) {
	err := sendToBrowser(pid)
	if err != nil {
		log.Printf("could not notify client %v, %v", pid, err)
	}
}

func sendToBrowser(pid uuid.UUID) error {
	lock.Lock()
	conns := clients[pid]
	lock.Unlock()

	if conns == nil {
		return fmt.Errorf("cannot get websockt for clients %v", pid)
	}

	for _, conn := range conns {
		err := conn.WriteJSON(SocketMessage{MessageType: "currencies", Data: enum.GetCryptoCurrencyDetails()})
		if err != nil {
			conn.Close()
		} else {
			clientsCopy[pid] = append(clientsCopy[pid], conn)
		}
	}
	clients[pid] = clientsCopy[pid]
	clientsCopy = make(map[uuid.UUID][]*websocket.Conn)

	for {
		for _, conn := range conns {
			var message SocketMessage
			err := conn.ReadJSON(message)
			if err != nil {
				log.Println("read failed:", err)
				break
			}
			cs := message.Data.(CurrencySelection)
			log.Println("cs", cs)
		}

	}

	go sendMessage()
	go sendMessage2()
	go sendMessage3()

	return nil
}

func sendMessage() {
	time.Sleep(5 * time.Second)
	pid := uuid.MustParse("a6e2b1bc-5d17-40d5-ae91-9cce9a8304b5")
	conns := clients[pid]
	for _, conn := range conns {
		err := conn.WriteJSON(SocketMessage{MessageType: "wait-for-tx", Data: enum.GetCryptoCurrencyDetails()})
		if err != nil {
			conn.Close()
		} else {
			clientsCopy[pid] = append(clientsCopy[pid], conn)
		}
	}
	clients[pid] = clientsCopy[pid]
	clientsCopy = make(map[uuid.UUID][]*websocket.Conn)
}

func sendMessage2() {
	time.Sleep(10 * time.Second)
	pid := uuid.MustParse("a6e2b1bc-5d17-40d5-ae91-9cce9a8304b5")
	conns := clients[pid]
	for _, conn := range conns {
		err := conn.WriteJSON(SocketMessage{MessageType: "received-tx", Data: enum.GetCryptoCurrencyDetails()})
		if err != nil {
			conn.Close()
		} else {
			clientsCopy[pid] = append(clientsCopy[pid], conn)
		}
	}
	clients[pid] = clientsCopy[pid]
	clientsCopy = make(map[uuid.UUID][]*websocket.Conn)
}

func sendMessage3() {
	time.Sleep(15 * time.Second)
	pid := uuid.MustParse("a6e2b1bc-5d17-40d5-ae91-9cce9a8304b5")
	conns := clients[pid]
	for _, conn := range conns {
		err := conn.WriteJSON(SocketMessage{MessageType: "confirmed", Data: enum.GetCryptoCurrencyDetails()})
		if err != nil {
			conn.Close()
		} else {
			clientsCopy[pid] = append(clientsCopy[pid], conn)
		}
	}
	clients[pid] = clientsCopy[pid]
	clientsCopy = make(map[uuid.UUID][]*websocket.Conn)
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}
