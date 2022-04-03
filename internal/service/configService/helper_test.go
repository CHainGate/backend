package configService

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/pkg/enum"

	"github.com/CHainGate/backend/configApi"
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/CHainGate/backend/proxyClientApi"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var jwtTest JwtTest
var testMerchant = &model.Merchant{
	FirstName: "Momo",
	LastName:  "",
	Email:     "momo@mail.com",
	Password:  []byte("test"),
	IsActive:  true,
	EmailVerification: model.EmailVerification{
		VerificationCode: 123456,
	},
}

type JwtTest struct {
	token  string
	issuer string
	exp    time.Time
}

type Interceptor struct {
	core         http.RoundTripper
	testFunction func(r *http.Request) (*http.Request, error)
}

func setup() {
	utils.NewOpts()
	utils.Opts.JwtSecret = "secret"
	utils.Opts.ApiKeySecret = "apiSecretKey1234"
	merchantId, err := uuid.Parse("b39310ec-59f9-454e-b1dd-2bcc18e9994f")
	if err != nil {
		panic(err)
	}
	testMerchant.ID = merchantId
	testMerchant.EmailVerification.MerchantId = merchantId
}

func findMerchantByEmailMock() repository.IMerchantRepository {
	mock, repo := NewMock()
	row := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "is_active", "created_at"}).
		AddRow(testMerchant.ID, testMerchant.FirstName, testMerchant.LastName, testMerchant.Email, testMerchant.Password, testMerchant.IsActive, testMerchant.CreatedAt)

	mock.ExpectQuery("SELECT").WithArgs(testMerchant.Email).WillReturnRows(row)
	return repo
}

func shutdown() {}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func NewMock() (sqlmock.Sqlmock, repository.IMerchantRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDb, err := gorm.Open(dialector, &gorm.Config{})
	merchantRepository, err := repository.NewMerchantRepository(gormDb)
	if err != nil {
		return nil, nil
	}
	return mock, merchantRepository
}

func TestCreateJwtToken(t *testing.T) {
	jwtDuration := time.Hour * 24
	token, err := createJwtToken("test@email.com", jwtDuration)
	if err != nil {
		t.Errorf("Cannot create JWT Token, got error %s", err.Error())
	}
	jwtTest.token = token
	jwtTest.issuer = "test@email.com"
	jwtTest.exp = time.Now().Add(time.Hour * 24)
}

func TestDecodeJwtToken(t *testing.T) {
	claims, err := decodeJwtToken(jwtTest.token)
	if err != nil {
		t.Errorf("Cannot decode JWT Token, got error %s", err.Error())
	}

	if claims.Issuer != jwtTest.issuer {
		t.Errorf("Expected issuer %s, but got %s", jwtTest.issuer, claims.Issuer)
	}

	// remove milliseconds. If time.now is mocked this is not needed anymore
	if claims.ExpiresAt.Time.Truncate(time.Second) != jwtTest.exp.Truncate(time.Second) {
		t.Errorf("Expected issuer %s, but got %s", jwtTest.exp, claims.ExpiresAt.Time)
	}
}

func TestGetMerchantByEmail(t *testing.T) {
	repo := findMerchantByEmailMock()
	merchant, err := getMerchantByEmail(testMerchant.Email, repo)
	if err != nil {
		t.Fatalf("Cannot find merchant by email, got error %s", err.Error())
	}
	if merchant.Email != testMerchant.Email {
		t.Errorf("Expected email %s, but got %s", testMerchant.Email, merchant.Email)
	}
}

func TestHandleAuthorization(t *testing.T) {
	repo := findMerchantByEmailMock()
	token, err := createJwtToken(testMerchant.Email, time.Hour*1)
	if err != nil {
		t.Errorf("")
	}
	bearer := "bearer " + token
	merchant, err := handleAuthorization(bearer, repo)
	if err != nil {
		t.Fatalf("handleAuthorization: got error %s", err.Error())
	}

	if merchant.Email != testMerchant.Email {
		t.Errorf("Expected email %s, but got %s", testMerchant.Email, merchant.Email)
	}
}

func TestCreateVerificationCode(t *testing.T) {
	const verificationCodeLength = 6
	code, err := createVerificationCode()
	if err != nil {
		t.Fatalf("createVerificationCode: got error %s", err.Error())
	}
	if len(code.String()) != verificationCodeLength {
		t.Errorf("Expected verification code with length of %d, but got %d", verificationCodeLength, len(code.String()))
	}
}

func TestHandleVerification(t *testing.T) {
	merchantId, _ := uuid.Parse("b39310ec-59f9-454e-b1dd-2bcc18e9994f")
	codeId, _ := uuid.Parse("b39310ec-59f9-454e-b1dd-000000000000")
	merchant := model.Merchant{
		Base:      model.Base{ID: merchantId},
		FirstName: "hans",
		LastName:  "meier",
		Password:  []byte("pw"),
		Email:     "test@mail.com",
		IsActive:  true,
		EmailVerification: model.EmailVerification{
			Base:             model.Base{ID: codeId},
			MerchantId:       merchantId,
			VerificationCode: 123456,
		},
	}
	mock, repo := NewMock()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"merchants\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("INSERT INTO \"email_verifications\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), merchantId, 123456, codeId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	mock.ExpectCommit()

	err := handleVerification(&merchant, 123456, repo)
	if err != nil {
		t.Error(err)
	}

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Expectations were not met: %s", err.Error())
	}

}

func TestCreateMerchant(t *testing.T) {
	merchantId := uuid.New()
	verificationCode := big.NewInt(123456)
	encryptedPassword := []byte("password")
	registerRequest := configApi.RegisterRequestDto{
		Email:     "hans@mail.ch",
		Password:  "password",
		FirstName: "hans",
		LastName:  "meier",
	}

	mock, repo := NewMock()
	mock.ExpectBegin()

	mock.ExpectQuery("INSERT INTO \"merchants\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), registerRequest.FirstName, registerRequest.LastName, registerRequest.Email, sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(merchantId))

	mock.ExpectQuery("INSERT INTO \"email_verifications\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), merchantId, verificationCode.Int64()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	mock.ExpectCommit()

	merchant, err := createMerchant(verificationCode, registerRequest, encryptedPassword, repo)
	if err != nil {
		t.Fatalf("Error occured during createMerchant: %s", err.Error())
	}

	if merchant.Email != registerRequest.Email {
		t.Errorf("Expected email %s, but got %s", registerRequest.Email, merchant.Email)
	}
}

func TestSendVerificationEmail(t *testing.T) {
	test := func(r *http.Request) (*http.Request, error) {
		expected := proxyClientApi.EmailRequestDto{
			Name:    "Momo",
			EmailTo: "momo@mail.com",
			Subject: "Verify your E-Mail",
			Content: "Please Verify your E-Mail: ?email=momo@mail.com&code=123456",
		}

		var actually proxyClientApi.EmailRequestDto
		err := json.NewDecoder(r.Body).Decode(&actually)
		if err != nil {
			return nil, err
		}

		if expected != actually {
			msg := fmt.Sprintf("Wrong request body. expected %v, but got %v", expected, actually)
			return nil, errors.New(msg)
		}

		return nil, nil
	}
	httpClient := http.Client{
		Transport: Interceptor{
			core:         http.DefaultTransport,
			testFunction: test,
		},
	}

	err := sendVerificationEmail(testMerchant, &httpClient)
	if err != nil {
		t.Error(err)
	}
}

// TODO: improve test
func TestHandleSecretApiKey(t *testing.T) {
	key, _, err := handleSecretApiKey("trfertfw3", enum.Test, enum.Secret)
	if err != nil {
		t.Fatal(err)
	}
	if key.Mode != enum.Test ||
		key.KeyType != enum.Secret ||
		key.IsActive != true {
		t.Errorf("")
	}
}

// TODO: improve test
func TestHandlePublicApiKey(t *testing.T) {
	key, err := handlePublicApiKey("trfertfw3", enum.Test, enum.Public)
	if err != nil {
		t.Fatal(err)
	}
	if key.Mode != enum.Test ||
		key.KeyType != enum.Public ||
		key.IsActive != true {
		t.Errorf("")
	}
}

func TestGetCombinedApiKey(t *testing.T) {
	secretKey := "supersecret"
	key := model.ApiKey{}
	key.ID = uuid.New()
	combinedApiKey, err := getCombinedApiKey(key, secretKey)
	if err != nil {
		t.Fatalf("getCombinedApiKey error: %s", err.Error())
	}
	decrypt, err := utils.Decrypt([]byte(utils.Opts.ApiKeySecret), combinedApiKey)
	if err != nil {
		t.Fatalf("Decrypt error: %s", err.Error())
	}

	expected := key.ID.String() + "_" + secretKey
	if expected != decrypt {
		t.Errorf("expected combined key %s, but got %s", expected, decrypt)
	}
}

func TestGetApiKeyHint(t *testing.T) {
	key := "lkja4j5lkjalfj235w4lbvsst"
	expected := "lkja...vsst"
	hint := getApiKeyHint(key)
	if hint != expected {
		t.Errorf("expected hint %s, but got %s", expected, hint)
	}
}

func (i Interceptor) RoundTrip(r *http.Request) (*http.Response, error) {
	defer func() {
		_ = r.Body.Close()
	}()
	_, err := i.testFunction(r)

	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK}, nil
}
