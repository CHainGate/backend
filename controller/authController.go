package controller

import (
	"CHainGate/backend/database"
	"CHainGate/backend/model"
	"CHainGate/backend/utils"
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := model.User{
		Name:     data["name"],
		Email:    data["email"],
		Password: password,
	}

	database.DB.Create(&user)

	if user.Id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user model.User
	database.DB.Where("email = ?", data["email"]).First(&user)

	if user.Id == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"]))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    strconv.Itoa(int(user.Id)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	})

	token, err := claims.SignedString([]byte(utils.Opts.JwtSecret))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cookie := &http.Cookie{
		Name:  "token",
		Value: token,
		//Secure: true, HTTPS only on PROD
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 24),
	}
	http.SetCookie(w, cookie)

	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func User(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.Opts.JwtSecret), nil
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	var user model.User

	database.DB.Where("id = ?", claims.Issuer).First(&user)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func Logout(w http.ResponseWriter, _ *http.Request) {
	cookie := &http.Cookie{
		Name:  "token",
		Value: "",
		//Secure: true, HTTPS only on PROD
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}
