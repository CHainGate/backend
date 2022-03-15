package main

import (
	"CHainGate/backend/configApi"
	"CHainGate/backend/internalApi"
	"CHainGate/backend/publicApi"
	"CHainGate/backend/service/configService"
	"CHainGate/backend/service/internalService"
	"CHainGate/backend/service/publicService"
	"CHainGate/backend/utils"
	"log"
	"net/http"
	"strconv"
)

func main() {
	utils.NewOpts() // create utils.Opts (env variables)
	//database.Connect()

	// config api
	ApiKeyApiService := configService.NewApiKeyApiService()
	ApiKeyApiController := configApi.NewApiKeyApiController(ApiKeyApiService)

	AuthenticationApiService := configService.NewAuthenticationApiService()
	AuthenticationApiController := configApi.NewAuthenticationApiController(AuthenticationApiService)

	LoggingApiService := configService.NewLoggingApiService()
	LoggingApiController := configApi.NewLoggingApiController(LoggingApiService)

	WalletApiService := configService.NewWalletApiService()
	WalletApiController := configApi.NewWalletApiController(WalletApiService)

	configRouter := configApi.NewRouter(ApiKeyApiController, AuthenticationApiController, LoggingApiController, WalletApiController)

	// public api
	PaymentApiService := publicService.NewPaymentApiService()
	PaymentApiController := publicApi.NewPaymentApiController(PaymentApiService)

	publicRouter := publicApi.NewRouter(PaymentApiController)

	// internal api
	PaymentUpdateApiService := internalService.NewPaymentUpdateApiService()
	PaymentUpdateApiController := internalApi.NewPaymentUpdateApiController(PaymentUpdateApiService)

	internalRouter := internalApi.NewRouter(PaymentUpdateApiController)

	http.Handle("/api/config/", configRouter)
	http.Handle("/api/public/", publicRouter)
	http.Handle("/api/internal/", internalRouter)

	// https://ribice.medium.com/serve-swaggerui-within-your-golang-application-5486748a5ed4
	configFs := http.FileServer(http.Dir("./swaggerui/config"))
	http.Handle("/api/config/swaggerui/", http.StripPrefix("/api/config/swaggerui/", configFs))
	publicFs := http.FileServer(http.Dir("./swaggerui/public"))
	http.Handle("/api/public/swaggerui/", http.StripPrefix("/api/public/swaggerui/", publicFs))
	internalFs := http.FileServer(http.Dir("./swaggerui/internal"))
	http.Handle("/api/internal/swaggerui/", http.StripPrefix("/api/internal/swaggerui/", internalFs))

	log.Println("Starting proxy-service on port " + strconv.Itoa(utils.Opts.ServerPort))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(utils.Opts.ServerPort), nil))
}
