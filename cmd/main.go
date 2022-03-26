package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/database"
	"github.com/CHainGate/backend/internal/service/configService"
	"github.com/CHainGate/backend/internal/service/internalService"
	"github.com/CHainGate/backend/internal/service/publicService"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/CHainGate/backend/internalApi"
	"github.com/CHainGate/backend/publicApi"

	"github.com/rs/cors"
)

func main() {
	utils.NewOpts() // create utils.Opts (env variables)
	database.Connect()

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

	log.Println("Starting backend-service on port " + strconv.Itoa(utils.Opts.ServerPort))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(utils.Opts.ServerPort), nil))
}
