package utils

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	Opts *OptsType
)

type OptsType struct {
	ServerPort           int
	DbHost               string
	DbUser               string
	DbPassword           string
	DbName               string
	DbPort               string
	JwtSecret            string
	SendGridApiKey       string
	EmailFrom            string
	EmailVerificationUrl string
}

func NewOpts() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &OptsType{}
	flag.IntVar(&o.ServerPort, "SERVER_PORT", lookupEnvInt("SERVER_PORT", 8000), "Server PORT")
	flag.StringVar(&o.DbHost, "DB_HOST", lookupEnv("DB_HOST"), "Database Host")
	flag.StringVar(&o.DbUser, "DB_USER", lookupEnv("DB_USER"), "Database User")
	flag.StringVar(&o.DbPassword, "DB_PASSWORD", lookupEnv("DB_PASSWORD"), "Database Password")
	flag.StringVar(&o.DbName, "DB_NAME", lookupEnv("DB_NAME"), "Database Name")
	flag.StringVar(&o.DbPort, "DB_PORT", lookupEnv("DB_PORT"), "Database Port")
	flag.StringVar(&o.JwtSecret, "JWT_SECRET", lookupEnv("JWT_SECRET"), "JWT Secret")
	flag.StringVar(&o.SendGridApiKey, "SENDGRID_API_KEY", lookupEnv("SENDGRID_API_KEY"), "SendGrid API Key")
	flag.StringVar(&o.EmailFrom, "EMAIL_FROM", lookupEnv("EMAIL_FROM"), "Email From")
	flag.StringVar(&o.EmailVerificationUrl, "EMAIL_VERIFICATION_URL", lookupEnv("EMAIL_VERIFICATION_URL"), "Email Verification URL")

	Opts = o
}

func lookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			return v
		}
	}
	return ""
}

func lookupEnvInt(key string, defaultValues ...int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("LookupEnvInt[%s]: %v", key, err)
			return 0
		}
		return v
	}
	for _, v := range defaultValues {
		if v != 0 {
			return v
		}
	}
	return 0
}
