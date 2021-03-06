package utils

import (
	"errors"
	"flag"
	"github.com/CHainGate/backend/pkg/enum"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type OptsType struct {
	ServerPort           int
	DbHost               string
	DbUser               string
	DbPassword           string
	DbName               string
	DbPort               string
	JwtSecret            string
	ApiKeySecret         string
	EmailVerificationUrl string
	ProxyBaseUrl         string
	EthereumBaseUrl      string
	BitcoinBaseUrl       string
	PaymentBaseUrl       string
}

var (
	Opts *OptsType
)

func NewOpts() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &OptsType{}
	//TODO: add default values
	flag.IntVar(&o.ServerPort, "SERVER_PORT", lookupEnvInt("SERVER_PORT", 8000), "Server PORT")
	flag.StringVar(&o.DbHost, "DB_HOST", lookupEnv("DB_HOST"), "Database Host")
	flag.StringVar(&o.DbUser, "DB_USER", lookupEnv("DB_USER"), "Database User")
	flag.StringVar(&o.DbPassword, "DB_PASSWORD", lookupEnv("DB_PASSWORD"), "Database Password")
	flag.StringVar(&o.DbName, "DB_NAME", lookupEnv("DB_NAME"), "Database Name")
	flag.StringVar(&o.DbPort, "DB_PORT", lookupEnv("DB_PORT"), "Database Port")
	flag.StringVar(&o.JwtSecret, "JWT_SECRET", lookupEnv("JWT_SECRET"), "JWT Secret")
	flag.StringVar(&o.ApiKeySecret, "API_KEY_SECRET", lookupEnv("API_KEY_SECRET"), "API Key Secret")
	flag.StringVar(&o.EmailVerificationUrl, "EMAIL_VERIFICATION_URL", lookupEnv("EMAIL_VERIFICATION_URL"), "Email Verification URL")
	flag.StringVar(&o.ProxyBaseUrl, "PROXY_BASE_URL", lookupEnv("PROXY_BASE_URL", "http://localhost:8001/api"), "Proxy base url")
	flag.StringVar(&o.EthereumBaseUrl, "ETHEREUM_BASE_URL", lookupEnv("ETHEREUM_BASE_URL", "http://localhost:9000/api"), "Ethereum base url")
	flag.StringVar(&o.BitcoinBaseUrl, "BITCOIN_BASE_URL", lookupEnv("BITCOIN_BASE_URL", "http://localhost:9001/api"), "Bitcoin base url")
	flag.StringVar(&o.PaymentBaseUrl, "PAYMENT_URL", lookupEnv("PAYMENT_URL", "http://localhost:3000/payment/"), "Payment base URL")

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

func ConvertAmountToBaseString(currency enum.CryptoCurrency, amount big.Int) (string, error) {
	if amount.Cmp(big.NewInt(0)) == 0 {
		return "0", nil
	}

	details := enum.GetCryptoCurrencyDetails()
	for _, c := range details {
		if currency.String() == c.ShortName {
			conversionFactor, err := strconv.ParseFloat(c.ConversionFactor, 64)
			if err != nil {
				return "", err
			}

			f := new(big.Float)
			f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
			f.SetMode(big.ToNearestEven)
			fWei := new(big.Float)
			fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
			fWei.SetMode(big.ToNearestEven)
			value := f.Quo(fWei.SetInt(&amount), big.NewFloat(conversionFactor))

			var precision int
			if currency == enum.ETH {
				precision = 18
			} else if currency == enum.BTC {
				precision = 8
			}
			return value.Text('f', precision), nil
		}
	}
	return "", errors.New("convertToBase amount failed")
}
