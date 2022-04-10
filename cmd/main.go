package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/CHainGate/backend/internal/service"

	"github.com/CHainGate/backend/internal/repository"

	"github.com/CHainGate/backend/configApi"
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

	log.Println("Starting backend-service on port " + strconv.Itoa(utils.Opts.ServerPort))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(utils.Opts.ServerPort), nil))
}
