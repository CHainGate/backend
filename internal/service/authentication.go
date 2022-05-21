package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/CHainGate/backend/proxyClientApi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const jwtDuration = time.Hour * 24

type IAuthenticationService interface {
	HandleJwtAuthentication(bearer string) (*model.Merchant, error)
	HandleLogin(email string, password string) (string, error)
	HandleApiAuthentication(apiKey string) (*model.Merchant, *model.ApiKey, error)
	CreateApiKey(mode enum.Mode) (*model.ApiKey, error)
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
	decryptedApiKey, err := Decrypt([]byte(utils.Opts.ApiKeySecret), apiKey)
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

	encryptedKeySecret, err := scryptPassword(apiKeySecret, currentApiKey.SecretSalt)
	if err != nil {
		return nil, nil, err
	}

	if encryptedKeySecret != currentApiKey.Secret {
		return nil, nil, errors.New("not authorized")
	}

	merchant, err := s.merchantRepository.FindById(currentApiKey.MerchantId)
	if err != nil {
		return nil, nil, err
	}

	return merchant, currentApiKey, nil
}

func (s *authenticationService) CreateApiKey(mode enum.Mode) (*model.ApiKey, error) {
	apiKeySecret, err := generateApiKeySecret()
	if err != nil {
		return nil, err
	}

	key := model.ApiKey{
		Base: model.Base{ID: uuid.New()},
		Mode: mode,
	}

	secretSalt, err := createSalt()
	if err != nil {
		return nil, err
	}

	apiKeySecretEncrypted, err := scryptPassword(apiKeySecret, secretSalt)
	if err != nil {
		return nil, err
	}

	key.Secret = apiKeySecretEncrypted
	key.SecretSalt = secretSalt

	combinedApiKey, err := getCombinedApiKey(key, apiKeySecret)
	if err != nil {
		return nil, err
	}

	encryptedCombinedApiKey, err := encrypt([]byte(utils.Opts.ApiKeySecret), combinedApiKey)
	if err != nil {
		return nil, err
	}

	key.ApiKey = encryptedCombinedApiKey

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
	salt, err := createSalt()
	if err != nil {
		return err
	}

	encryptedPassword, err := scryptPassword(registerRequestDto.Password, salt)
	if err != nil {
		return err
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
		Salt:              salt,
		EmailVerification: emailVerification,
		IsActive:          false,
	}

	err = sendVerificationEmail(&merchant)
	if err != nil {
		return err
	}

	err = s.merchantRepository.Create(&merchant)
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

func sendVerificationEmail(merchant *model.Merchant) error {
	baseUrl, err := url.Parse(utils.Opts.EmailVerificationUrl)
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Add("email", merchant.Email)
	params.Add("code", strconv.FormatUint(merchant.EmailVerification.VerificationCode, 10))

	baseUrl.RawQuery = params.Encode()

	content := "Please Verify your E-Mail: " + baseUrl.String()
	email := *proxyClientApi.NewEmailRequestDto(merchant.FirstName, merchant.Email, "Verify your E-Mail", content)
	configuration := proxyClientApi.NewConfiguration()
	configuration.Servers[0].URL = utils.Opts.ProxyBaseUrl
	apiClient := proxyClientApi.NewAPIClient(configuration)
	_, err = apiClient.EmailApi.SendEmail(context.Background()).EmailRequestDto(email).Execute()
	if err != nil {
		return err
	}
	return nil
}

func canMerchantLogin(merchant *model.Merchant, password string) error {
	if !merchant.IsActive {
		return errors.New("Merchant not active ")
	}
	encryptedPassword, err := scryptPassword(password, merchant.Salt)
	if err != nil {
		return err
	}
	if encryptedPassword != merchant.Password {
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
