package main

import (
	"CHainGate/backend/database"
	"CHainGate/backend/openapi"
	"CHainGate/backend/service"
	"CHainGate/backend/utils"
	"log"
	"net/http"
)

func main() {
	utils.NewOpts() // create utils.Opts (env variables)
	database.Connect()

	PetApiService := service.NewPetApiService()
	PetApiController := openapi.NewPetApiController(PetApiService)

	StoreApiService := service.NewStoreApiService()
	StoreApiController := openapi.NewStoreApiController(StoreApiService)

	UserApiService := service.NewUserApiService()
	UserApiController := openapi.NewUserApiController(UserApiService)

	router := openapi.NewRouter(PetApiController, StoreApiController, UserApiController)

	// https://ribice.medium.com/serve-swaggerui-within-your-golang-application-5486748a5ed4
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("./swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)

	log.Fatal(http.ListenAndServe(":9000", router))

	//routes.Setup()
}
