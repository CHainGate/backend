package controller

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

func CreateSecureApiKey(w http.ResponseWriter, _ *http.Request) {
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
}
