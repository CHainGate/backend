package enum

import "strings"

type CryptoCurrency int

// https://levelup.gitconnected.com/implementing-enums-in-golang-9537c433d6e2
const (
	ETH CryptoCurrency = iota + 1
	BTC
)

func (f CryptoCurrency) String() string {
	return [...]string{"eth", "btc"}[f-1]
}

// ParseStringToCryptoCurrencyEnum https://stackoverflow.com/questions/68543604/best-way-to-parse-a-string-to-an-enum
func ParseStringToCryptoCurrencyEnum(str string) (CryptoCurrency, bool) {
	capabilitiesMap := map[string]CryptoCurrency{
		"eth": ETH,
		"btc": BTC,
	}
	c, ok := capabilitiesMap[strings.ToLower(str)]
	return c, ok
}

func GetCryptoCurrencyDetails() map[string]string {
	details := map[string]string{
		"eth": "Ethereum",
		"btc": "Bitcoin",
	}
	return details
}
