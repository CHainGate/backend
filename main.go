package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/test", test).Methods("GET", "OPTIONS")
	log.Printf("listing on port %v", 8000)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(8000), router))
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Test:")
	err := json.NewEncoder(w).Encode("Test")
	if err != nil {
		return
	}
	return
}
