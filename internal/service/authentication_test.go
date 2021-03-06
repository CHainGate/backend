package service

import (
	"log"
	"os"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/internal/repository"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var service IAuthenticationService
var mock sqlmock.Sqlmock
var jwtTest JwtTest
var testMerchant = &model.Merchant{
	FirstName: "Momo",
	LastName:  "",
	Email:     "momo@mail.com",
	Password:  "test",
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

func setup() {
	newMock, merchantRepo := NewMerchantRepositoryMock()
	mock = newMock
	_, apiKeyRepo := NewApiKeyRepositoryMock()
	service = NewAuthenticationService(merchantRepo, apiKeyRepo)
	utils.NewOpts()
	utils.Opts.JwtSecret = "secret"
	utils.Opts.ApiKeySecret = "apiSecretKey1234"
	merchantId, err := uuid.Parse("b39310ec-59f9-454e-b1dd-2bcc18e9994f")
	if err != nil {
		log.Fatal(err)
	}
	testMerchant.ID = merchantId
	testMerchant.EmailVerification.MerchantId = merchantId
}

/*func findMerchantByEmailMock() repository.IMerchantRepository {
	row := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "is_active", "created_at"}).
		AddRow(testMerchant.ID, testMerchant.FirstName, testMerchant.LastName, testMerchant.Email, testMerchant.Password, testMerchant.IsActive, testMerchant.CreatedAt)

	merchantMock.ExpectQuery("SELECT").WithArgs(testMerchant.Email).WillReturnRows(row)
	return repo
}*/

func shutdown() {}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func NewMerchantRepositoryMock() (sqlmock.Sqlmock, repository.IMerchantRepository) {
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

func NewApiKeyRepositoryMock() (sqlmock.Sqlmock, repository.IApiKeyRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDb, err := gorm.Open(dialector, &gorm.Config{})
	apiKeyRepository, err := repository.NewApiKeyRepository(gormDb)
	if err != nil {
		return nil, nil
	}
	return mock, apiKeyRepository
}

func TestCreateJwtToken(t *testing.T) {
	jwtDuration := time.Hour * 24
	token, err := createJwtToken("test@email.com", "test", jwtDuration)
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

func TestHandleAuthorization(t *testing.T) {
	merchantRow := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "is_active", "created_at"}).
		AddRow(testMerchant.ID, testMerchant.FirstName, testMerchant.LastName, testMerchant.Email, testMerchant.Password, testMerchant.IsActive, testMerchant.CreatedAt)
	verificationRow := sqlmock.NewRows([]string{"id", "merchant_id", "verification_code"}).AddRow(uuid.New(), testMerchant.ID, 123456)

	mock.ExpectQuery("SELECT (.+) FROM \"merchants\"").WithArgs(testMerchant.Email).WillReturnRows(merchantRow)
	mock.ExpectQuery("SELECT (.+) FROM \"email_verifications\"").WithArgs(testMerchant.ID).WillReturnRows(verificationRow)
	mock.ExpectQuery("SELECT (.+) FROM \"wallets\"").WithArgs(testMerchant.ID).WillReturnRows(sqlmock.NewRows([]string{""}))

	token, err := createJwtToken(testMerchant.Email, testMerchant.FirstName, time.Hour*1)
	if err != nil {
		t.Errorf("")
	}
	bearer := "bearer " + token
	merchant, err := service.HandleJwtAuthentication(bearer)
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
		Password:  "pw",
		Salt:      []byte("salt"),
		Email:     "test@mail.com",
		IsActive:  true,
		EmailVerification: model.EmailVerification{
			Base:             model.Base{ID: codeId},
			MerchantId:       merchantId,
			VerificationCode: 123456,
		},
	}

	merchantRow := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "salt", "is_active", "created_at"}).
		AddRow(merchant.ID, merchant.FirstName, merchant.LastName, merchant.Email, merchant.Password, merchant.Salt, merchant.IsActive, merchant.CreatedAt)
	verificationRow := sqlmock.NewRows([]string{"id", "merchant_id", "verification_code"}).AddRow(merchant.EmailVerification.ID, testMerchant.ID, merchant.EmailVerification.VerificationCode)

	mock.ExpectQuery("SELECT (.+) FROM \"merchants\"").WithArgs(merchant.Email).WillReturnRows(merchantRow)
	mock.ExpectQuery("SELECT (.+) FROM \"email_verifications\"").WithArgs(merchant.ID).WillReturnRows(verificationRow)
	mock.ExpectQuery("SELECT (.+) FROM \"wallets\"").WithArgs(testMerchant.ID).WillReturnRows(sqlmock.NewRows([]string{""}))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"merchants\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("INSERT INTO \"email_verifications\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), merchantId, 123456, codeId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	mock.ExpectCommit()

	err := service.HandleVerification(merchant.Email, 123456)
	if err != nil {
		t.Error(err)
	}

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Expectations were not met: %s", err.Error())
	}

}

//TODO: make test work
/*func TestCreateMerchant(t *testing.T) {
	request := configApi.RegisterRequestDto{
		FirstName: "",
		LastName: "",
		Email: "",
		Password: "",
	}
	merchantId := uuid.New()
	verificationCode := big.NewInt(123456)
	encryptedPassword := []byte("password")
	registerRequest := configApi.RegisterRequestDto{
		Email:     "hans@mail.ch",
		Password:  "password",
		FirstName: "hans",
		LastName:  "meier",
	}


	merchantMock.ExpectBegin()

	merchantMock.ExpectQuery("INSERT INTO \"merchants\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), registerRequest.FirstName, registerRequest.LastName, registerRequest.Email, sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(merchantId))

	merchantMock.ExpectQuery("INSERT INTO \"email_verifications\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), merchantId, verificationCode.Int64()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	merchantMock.ExpectCommit()

	err := service.CreateMerchant(request)
	if err != nil {
		t.Fatalf("Error occured during createMerchant: %s", err.Error())
	}
}*/

func TestSendVerificationEmail(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	gock.New("localhost:8001").
		Post("/api/email").
		MatchType("json").
		JSON(map[string]string{
			"name":     "Momo",
			"email_to": "momo@mail.com",
			"subject":  "Verify your E-Mail",
			"content":  "Please Verify your E-Mail: ?code=123456&email=momo%40mail.com"}).
		Reply(200)

	err := sendVerificationEmail(testMerchant)
	if err != nil {
		t.Error(err)
	}
}

// TODO: improve test
func TestHandleSecretApiKey(t *testing.T) {
	key, err := service.CreateApiKey(enum.Test)
	if err != nil {
		t.Fatal(err)
	}
	if key.Mode != enum.Test {
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
	decrypt, err := Decrypt([]byte(utils.Opts.ApiKeySecret), combinedApiKey)
	if err != nil {
		t.Fatalf("Decrypt error: %s", err.Error())
	}

	expected := key.ID.String() + "_" + secretKey
	if expected != decrypt {
		t.Errorf("expected combined key %s, but got %s", expected, decrypt)
	}
}
