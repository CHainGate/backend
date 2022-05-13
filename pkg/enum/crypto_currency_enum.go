package enum

import "strings"

type CryptoCurrency int
type Currency struct {
	Name             string
	ShortName        string
	ConversionFactor string
}

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

func GetCryptoCurrencyDetails() []Currency {
	return []Currency{
		{"Ethereum", "eth", "1000000000000000000"},
		{"Bitcoin", "btc", "100000000"},
	}
}
