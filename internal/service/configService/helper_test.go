package configService

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/CHainGate/backend/internal/models"
	"github.com/CHainGate/backend/internal/repository/userRepository"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var jwtTest JwtTest
var u = &models.User{
	FirstName: "Momo",
	LastName:  "",
	Email:     "momo@mail.com",
	Password:  []byte("test"),
	IsActive:  true,
	CreatedAt: time.Now(),
	EmailVerification: models.EmailVerification{
		Id:               uuid.New(),
		VerificationCode: 123456,
		CreatedAt:        time.Now(),
	},
}

type JwtTest struct {
	token  string
	issuer string
	exp    time.Time
}

func setup() {
	utils.NewOpts()
	utils.Opts.JwtSecret = "secret"
	userId, err := uuid.Parse("b39310ec-59f9-454e-b1dd-2bcc18e9994f")
	if err != nil {
		panic(err)
	}
	u.Id = userId
	u.EmailVerification.UserId = userId
}

func findUserByEmailMock() userRepository.IUserRepository {
	mock, repo := NewMock()
	row := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "is_active", "created_at"}).
		AddRow(u.Id, u.FirstName, u.LastName, u.Email, u.Password, u.IsActive, u.CreatedAt)

	mock.ExpectQuery("SELECT").WithArgs(u.Email).WillReturnRows(row)
	return repo
}

func shutdown() {}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func NewMock() (sqlmock.Sqlmock, userRepository.IUserRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDb, err := gorm.Open(dialector, &gorm.Config{})
	return mock, &userRepository.UserRepository{DB: gormDb}
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

func TestGetUserByEmail(t *testing.T) {
	repo := findUserByEmailMock()
	user, err := getUserByEmail(u.Email, repo)
	if err != nil {
		t.Fatalf("Cannot find user by email, got error %s", err.Error())
	}
	if user.Email != u.Email {
		t.Errorf("Expected email %s, but got %s", u.Email, user.Email)
	}
}

func TestCheckAuthorizationAndReturnUser(t *testing.T) {
	repo := findUserByEmailMock()
	token, err := createJwtToken(u.Email, time.Hour*1)
	if err != nil {
		t.Errorf("")
	}
	bearer := "bearer " + token
	user, err := checkAuthorizationAndReturnUser(bearer, repo)
	if err != nil {
		t.Fatalf("checkAuthorizationAndReturnUser: got error %s", err.Error())
	}

	if user.Email != u.Email {
		t.Errorf("Expected email %s, but got %s", u.Email, user.Email)
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
	userId, _ := uuid.Parse("b39310ec-59f9-454e-b1dd-2bcc18e9994f")
	codeId, _ := uuid.Parse("b39310ec-59f9-454e-b1dd-000000000000")
	usr := models.User{
		Id:        userId,
		FirstName: "hans",
		LastName:  "meier",
		Password:  []byte("pw"),
		Email:     "test@mail.com",
		IsActive:  true,
		CreatedAt: time.Now(),
		EmailVerification: models.EmailVerification{
			Id:               codeId,
			UserId:           userId,
			VerificationCode: 123456,
			CreatedAt:        time.Now(),
		},
	}

	mock, repo := NewMock()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"users\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("INSERT INTO \"email_verifications\"").
		WithArgs(userId, 123456, sqlmock.AnyArg(), codeId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	mock.ExpectCommit()

	err := handleVerification(&usr, 123456, repo)
	if err != nil {
		t.Error(err)
	}

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Expectations were not met: %s", err.Error())
	}

}
