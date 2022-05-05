package main

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strconv"
	"sync"

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

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var lock = sync.Mutex{}
var clients = make(map[uuid.UUID][]*websocket.Conn)

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
	PaymentApiService := publicService.NewPaymentApiService(publicPaymentService, authService)
	PaymentApiController := publicApi.NewPaymentApiController(PaymentApiService)

	publicRouter := publicApi.NewRouter(PaymentApiController)

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

	lock.Lock()
	clients[user.Id] = conn
	lock.Unlock()
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("closing connection for %v", user.Id)
		lock.Lock()
		delete(clients, user.Id)
		lock.Unlock()
		return nil
	})

	conn.SetPongHandler(func(appData string) error {
		log.Printf(appData)
		return nil
	})

	notifyBrowser(user.Id, user.PaymentCycleInId)
}

func notifyBrowser(uid uuid.UUID, paymentCycleId *uuid.UUID) {
	go func(uid uuid.UUID, paymentCycleId *uuid.UUID) {
		err := sendToBrowser(uid, paymentCycleId)
		if err != nil {
			log.Warnf("could not notify client %v, %v", uid, err)
		}
	}(uid, paymentCycleId)
}

func sendToBrowser(userId uuid.UUID, paymentCycleInId *uuid.UUID) error {
	lock.Lock()
	conn := clients[userId]
	lock.Unlock()

	if conn == nil {
		return fmt.Errorf("cannot get websockt for client %v", userId)
	}

	userBalances, err := findUserBalances(userId)
	if err != nil {
		conn.Close()
		return err
	}

	err = conn.WriteJSON(UserBalances{PaymentCycle: pc, UserBalances: userBalancesDto, Total: total, DaysLeft: daysLeft})
	if err != nil {
		conn.Close()
		return err
	}

	return nil
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
