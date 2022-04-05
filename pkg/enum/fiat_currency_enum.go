package enum

import "strings"

type FiatCurrency int

// https://levelup.gitconnected.com/implementing-enums-in-golang-9537c433d6e2
const (
	USD FiatCurrency = iota + 1
	CHF
)

func (f FiatCurrency) String() string {
	return [...]string{"usd", "chf"}[f-1]
}

// ParseStringToFiatCurrencyEnum https://stackoverflow.com/questions/68543604/best-way-to-parse-a-string-to-an-enum
func ParseStringToFiatCurrencyEnum(str string) (FiatCurrency, bool) {
	capabilitiesMap := map[string]FiatCurrency{
		"usd": USD,
		"chf": CHF,
	}
	c, ok := capabilitiesMap[strings.ToLower(str)]
	return c, ok
}

func GetFiatCurrencyDetails() map[string]string {
	details := map[string]string{
		"usd": "US Dollar",
		"chf": "Swiss Franc",
	}
	return details
}
