package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/test", test).Methods("GET", "OPTIONS")
	router.HandleFunc("/apikey", createSecureApiKey).Methods("GET", "OPTIONS")
	log.Printf("listing on port %v", 8000)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(8000), router))
}

func test(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode("Test")
	if err != nil {
		return
	}
	return
}

func createSecureApiKey(w http.ResponseWriter, _ *http.Request) {
	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return
	}
	apiKey := base64.StdEncoding.EncodeToString(randomBytes)
	err = json.NewEncoder(w).Encode(apiKey)
	if err != nil {
		return
	}
	return
}
