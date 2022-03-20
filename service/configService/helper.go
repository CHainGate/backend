package configService

import (
	"CHainGate/backend/database"
	"CHainGate/backend/models"
	"CHainGate/backend/utils"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

func checkAuthorizationAndReturnUser(bearer string) (models.User, error) {
	bearerToken := strings.Split(bearer, " ")

	token, err := jwt.ParseWithClaims(bearerToken[1], &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.Opts.JwtSecret), nil
	})
	if err != nil {

	}
	claims := token.Claims.(*jwt.RegisteredClaims)

	var user models.User

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return models.User{}, errors.New("token expired")
	}
	result := database.DB.Where("email = ?", claims.Issuer).First(&user)
	if result.Error != nil {
		return models.User{}, result.Error
	}

	return user, nil
}
