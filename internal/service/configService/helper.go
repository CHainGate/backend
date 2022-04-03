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

	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/pkg/enum"

	"github.com/google/uuid"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/proxyClientApi"
	"golang.org/x/crypto/bcrypt"

	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/utils"

	"github.com/golang-jwt/jwt/v4"
)

func handleAuthorization(bearer string, repo repository.IMerchantRepository) (*model.Merchant, error) {
	bearerToken := strings.Split(bearer, " ")
	claims, err := decodeJwtToken(bearerToken[1])
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	merchant, err := getMerchantByEmail(claims.Issuer, repo)
	if err != nil {
		return nil, err
	}

	return merchant, nil
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

func getMerchantByEmail(email string, repo repository.IMerchantRepository) (*model.Merchant, error) {
	merchant, err := repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	return merchant, nil
}

//TODO bcrypt to scrypt and test
func canMerchantLogin(merchant *model.Merchant, password string) error {
	if !merchant.IsActive {
		return errors.New("Merchant not active ")
	}
	err := bcrypt.CompareHashAndPassword(merchant.Password, []byte(password))
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

func sendVerificationEmail(merchant *model.Merchant, client *http.Client) error {
	url := utils.Opts.EmailVerificationUrl + "?email=" + merchant.Email + "&code=" + strconv.FormatUint(merchant.EmailVerification.VerificationCode, 10)
	content := "Please Verify your E-Mail: " + url
	email := *proxyClientApi.NewEmailRequestDto(merchant.FirstName, merchant.Email, "Verify your E-Mail", content)
	configuration := proxyClientApi.NewConfiguration()
	configuration.HTTPClient = client
	apiClient := proxyClientApi.NewAPIClient(configuration)
	_, err := apiClient.EmailApi.SendEmail(context.Background()).EmailRequestDto(email).Execute()
	if err != nil {
		return errors.New("Verification E-Mail could not be sent ")
	}
	return nil
}

func createMerchant(
	verificationCode *big.Int,
	registerRequestDto configApi.RegisterRequestDto,
	encryptedPassword []byte,
	repo repository.IMerchantRepository,
) (*model.Merchant, error) {
	emailVerification := model.EmailVerification{
		VerificationCode: verificationCode.Uint64(),
	}

	merchant := model.Merchant{
		FirstName:         registerRequestDto.FirstName,
		LastName:          registerRequestDto.LastName,
		Email:             registerRequestDto.Email,
		Password:          encryptedPassword,
		EmailVerification: emailVerification,
		IsActive:          false,
	}

	err := repo.Create(&merchant)
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func handleVerification(merchant *model.Merchant, verificationCode int64, repo repository.IMerchantRepository) error {
	if merchant.EmailVerification.VerificationCode == uint64(verificationCode) {
		merchant.IsActive = true
		err := repo.Update(merchant)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Wrong verification code ")
}

func handleSecretApiKey(apiSecretKey string, mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, string, error) {
	key := model.ApiKey{
		Mode:     mode,
		KeyType:  apiKeyType,
		IsActive: true,
	}
	key.ID = uuid.New()

	salt, err := utils.CreateSalt()
	if err != nil {
		return nil, "", err
	}

	apiSecureKeyEncrypted, err := utils.ScryptPassword(apiSecretKey, salt)
	if err != nil {
		return nil, "", err
	}

	key.SecretKey = apiSecureKeyEncrypted
	key.Salt = salt

	combinedApiKey, err := getCombinedApiKey(key, apiSecretKey)
	if err != nil {
		return nil, "", err
	}

	key.ApiKey = getApiKeyHint(combinedApiKey)

	return &key, combinedApiKey, nil
}

func handlePublicApiKey(apiSecretKey string, mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, error) {
	key := model.ApiKey{
		Mode:     mode,
		KeyType:  apiKeyType,
		IsActive: true,
	}

	combinedApiKey, err := getCombinedApiKey(key, apiSecretKey)
	if err != nil {
		return nil, err
	}

	key.ApiKey = combinedApiKey
	key.SecretKey = apiSecretKey

	return &key, nil
}

func getCombinedApiKey(key model.ApiKey, apiSecretKey string) (string, error) {
	combinedKey := key.ID.String() + "_" + apiSecretKey
	return utils.Encrypt([]byte(utils.Opts.ApiKeySecret), combinedKey)
}

func getApiKeyHint(key string) string {
	apiKeyBeginning := key[0:4]
	apiKeyEnding := key[len(key)-4:]
	return apiKeyBeginning + "..." + apiKeyEnding // show the first and last 4 letters of the secret api key
}
