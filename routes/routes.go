package routes

import (
	"CHainGate/backend/controller"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func Setup() {
	router := mux.NewRouter()

	router.HandleFunc("/register", controller.Register).Methods("POST")
	router.HandleFunc("/login", controller.Login).Methods("POST")
	router.HandleFunc("/user", controller.User).Methods("GET")
	router.HandleFunc("/logout", controller.Logout).Methods("GET")

	router.HandleFunc("/apikey", controller.CreateSecureApiKey).Methods("GET", "OPTIONS")

	log.Printf("listing on port %v", 8000)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(8000), router))
}
