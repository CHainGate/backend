package service

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/CHainGate/backend/proxyClientApi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const jwtDuration = time.Hour * 24

type IAuthenticationService interface {
	HandleJwtAuthentication(bearer string) (*model.Merchant, error)
	HandleLogin(email string, password string) (string, error)
	HandleApiAuthentication(apiKey string) (*model.Merchant, *model.ApiKey, error)
	CreateSecretApiKey(mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, string, error)
	CreatePublicApiKey(mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, error)
	CreateMerchant(registerRequestDto configApi.RegisterRequestDto) error
	HandleVerification(email string, verificationCode int64) error
}

type authenticationService struct {
	merchantRepository repository.IMerchantRepository
	apiKeyRepository   repository.IApiKeyRepository
}

func NewAuthenticationService(
	merchantRepository repository.IMerchantRepository,
	apiKeyRepository repository.IApiKeyRepository,
) IAuthenticationService {
	return &authenticationService{merchantRepository, apiKeyRepository}
}

func (s *authenticationService) HandleJwtAuthentication(bearer string) (*model.Merchant, error) {
	bearerToken := strings.Split(bearer, " ")
	claims, err := decodeJwtToken(bearerToken[1])
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	merchant, err := s.merchantRepository.FindByEmail(claims.Issuer)
	if err != nil {
		return nil, err
	}

	return merchant, nil
}

func (s *authenticationService) HandleLogin(email string, password string) (string, error) {
	merchant, err := s.merchantRepository.FindByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.New("Email or password wrong ")
	}

	err = canMerchantLogin(merchant, password)
	if err != nil {
		return "", err
	}

	token, err := createJwtToken(merchant.Email, jwtDuration)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authenticationService) HandleApiAuthentication(apiKey string) (*model.Merchant, *model.ApiKey, error) {
	decryptedApiKey, err := decrypt([]byte(utils.Opts.ApiKeySecret), apiKey)
	if err != nil {
		return nil, nil, err
	}

	apiKeyDetails := strings.Split(decryptedApiKey, "_")
	apiKeyId := apiKeyDetails[0]
	apiKeySecret := apiKeyDetails[1]

	currentApiKey, err := s.apiKeyRepository.FindById(apiKeyId)
	if err != nil {
		return nil, nil, err
	}

	if currentApiKey.KeyType == enum.Secret {
		encryptedKey, err := scryptPassword(apiKeySecret, currentApiKey.Salt)
		if err != nil {
			return nil, nil, err
		}

		if encryptedKey != currentApiKey.SecretKey {
			return nil, nil, errors.New("not authorized")
		}
	}

	if currentApiKey.KeyType == enum.Public {
		if apiKeySecret != currentApiKey.SecretKey {
			return nil, nil, errors.New("not authorized")
		}
	}

	merchant, err := s.merchantRepository.FindById(currentApiKey.MerchantId)
	if err != nil {
		return nil, nil, err
	}

	return merchant, currentApiKey, nil
}

func (s *authenticationService) CreateSecretApiKey(mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, string, error) {
	apiSecretKey, err := generateApiKey()
	if err != nil {
		return nil, "", err
	}

	key := model.ApiKey{
		Mode:     mode,
		KeyType:  apiKeyType,
		IsActive: true,
	}
	key.ID = uuid.New()

	salt, err := createSalt()
	if err != nil {
		return nil, "", err
	}

	apiSecureKeyEncrypted, err := scryptPassword(apiSecretKey, salt)
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

func (s *authenticationService) CreatePublicApiKey(mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, error) {
	apiSecretKey, err := generateApiKey()
	if err != nil {
		return nil, err
	}
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

func (s *authenticationService) HandleVerification(email string, verificationCode int64) error {
	merchant, err := s.merchantRepository.FindByEmail(email)
	if err != nil {
		return err
	}

	if merchant.EmailVerification.VerificationCode == uint64(verificationCode) {
		merchant.IsActive = true
		err := s.merchantRepository.Update(merchant)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Wrong verification code ")
}

func (s *authenticationService) CreateMerchant(registerRequestDto configApi.RegisterRequestDto) error {
	//TODO: maybe use password validator https://github.com/wagslane/go-password-validator
	encryptedPassword, err := encryptPassword(registerRequestDto.Password)
	if err != nil {
		return errors.New("Cannot register merchant ")
	}

	verificationCode, err := createVerificationCode()
	if err != nil {
		return err
	}

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

	err = s.merchantRepository.Create(&merchant)
	if err != nil {
		return err
	}

	err = sendVerificationEmail(&merchant, nil)
	if err != nil {
		return err
	}

	return nil
}

func createJwtToken(issuer string, duration time.Duration) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
	})

	return claims.SignedString([]byte(utils.Opts.JwtSecret))
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

//TODO bcrypt to scrypt and test
func encryptPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

func getCombinedApiKey(key model.ApiKey, apiSecretKey string) (string, error) {
	combinedKey := key.ID.String() + "_" + apiSecretKey
	return encrypt([]byte(utils.Opts.ApiKeySecret), combinedKey)
}

func getApiKeyHint(key string) string {
	apiKeyBeginning := key[0:4]
	apiKeyEnding := key[len(key)-4:]
	return apiKeyBeginning + "..." + apiKeyEnding // show the first and last 4 letters of the secret api key
}
