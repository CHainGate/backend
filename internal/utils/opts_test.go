package utils

import (
	"os"
	"testing"
)

func TestNewOpts(t *testing.T) {
	expected := OptsType{
		ServerPort:           8000,
		DbHost:               "mydbhost",
		DbUser:               "postgres_usr",
		DbPassword:           "postgres_pw",
		DbName:               "postgres_db",
		DbPort:               "5432",
		JwtSecret:            "jwt_secret_token",
		ApiKeySecret:         "api_secret_key",
		EmailVerificationUrl: "https://send.email.ch/mail",
	}

	_ = os.Setenv("SERVER_PORT", "8000")
	_ = os.Setenv("DB_HOST", "mydbhost")
	_ = os.Setenv("DB_USER", "postgres_usr")
	_ = os.Setenv("DB_PASSWORD", "postgres_pw")
	_ = os.Setenv("DB_NAME", "postgres_db")
	_ = os.Setenv("DB_PORT", "5432")
	_ = os.Setenv("JWT_SECRET", "jwt_secret_token")
	_ = os.Setenv("API_KEY_SECRET", "api_secret_key")
	_ = os.Setenv("EMAIL_VERIFICATION_URL", "https://send.email.ch/mail")

	NewOpts()

	if expected != *Opts {
		t.Errorf("expected %v, but got %v", expected, Opts)
	}

}
