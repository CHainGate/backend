package configService

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/repository/userRepository"
	"github.com/CHainGate/backend/proxyClientApi"
	"golang.org/x/crypto/bcrypt"

	"github.com/CHainGate/backend/internal/models"
	"github.com/CHainGate/backend/internal/utils"

	"github.com/golang-jwt/jwt/v4"
)

func checkAuthorizationAndReturnUser(bearer string, repo userRepository.IUserRepository) (*models.User, error) {
	bearerToken := strings.Split(bearer, " ")
	claims, err := decodeJwtToken(bearerToken[1])
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	user, err := getUserByEmail(claims.Issuer, repo)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func decodeJwtToken(jwtToken string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.Opts.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	return claims, nil
}

func createJwtToken(issuer string, duration time.Duration) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
	})

	return claims.SignedString([]byte(utils.Opts.JwtSecret))
}

func getUserByEmail(email string, repo userRepository.IUserRepository) (*models.User, error) {
	user, err := repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

//TODO bcrypt to scrypt and test
func canUserLogin(user *models.User, password string) error {
	if !user.IsActive {
		return errors.New("User not active ")
	}
	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return errors.New("Wrong username or password ")
	}
	return nil
}

func createVerificationCode() (*big.Int, error) {
	// 6 digits
	max := big.NewInt(1000000)
	min := big.NewInt(100000)
	verificationCode, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		return nil, errors.New("Cannot generate verification code ")
	}
	verificationCode.Add(verificationCode, min)
	return verificationCode, nil
}

//TODO bcrypt to scrypt and test
func encryptPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func sendVerificationEmail(user models.User, client *http.Client) error {
	url := utils.Opts.EmailVerificationUrl + "?email=" + user.Email + "&code=" + strconv.FormatUint(user.EmailVerification.VerificationCode, 10)
	content := "Please Verify your E-Mail: " + url
	email := *proxyClientApi.NewEmailRequestDto(user.FirstName, user.Email, "Verify your E-Mail", content)
	configuration := proxyClientApi.NewConfiguration()
	configuration.HTTPClient = client
	apiClient := proxyClientApi.NewAPIClient(configuration)
	_, err := apiClient.EmailApi.SendEmail(context.Background()).EmailRequestDto(email).Execute()
	if err != nil {
		return errors.New("Verification E-Mail could not be sent ")
	}
	return nil
}

func createUser(
	verificationCode *big.Int,
	registerRequestDto configApi.RegisterRequestDto,
	encryptedPassword []byte,
	repo userRepository.IUserRepository,
) (models.User, error) {
	emailVerification := models.EmailVerification{
		VerificationCode: verificationCode.Uint64(),
		CreatedAt:        time.Now(),
	}

	user := models.User{
		FirstName:         registerRequestDto.FirstName,
		LastName:          registerRequestDto.LastName,
		Email:             registerRequestDto.Email,
		Password:          encryptedPassword,
		EmailVerification: emailVerification,
		IsActive:          false,
		CreatedAt:         time.Now(),
	}

	err := repo.CreateUser(&user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func handleVerification(user *models.User, verificationCode int64, repo userRepository.IUserRepository) error {
	if user.EmailVerification.VerificationCode == uint64(verificationCode) {
		user.IsActive = true
		err := repo.UpdateUser(user)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Wrong verification code ")
}
