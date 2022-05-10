package enum

import "strings"

type State int

// https://levelup.gitconnected.com/implementing-enums-in-golang-9537c433d6e2
const (
	CurrencySelection State = iota + 1
	Waiting
	PartiallyPaid
	Paid
	Confirmed
	Forwarded
	Finished
	Expired
	Failed
)

func (s State) String() string {
	return [...]string{"currency_selection", "waiting", "partially_paid", "paid", "confirmed", "forwarded", "finished", "expired", "failed"}[s-1]
}

// ParseStringToStateEnum ParseStringToModeEnum https://stackoverflow.com/questions/68543604/best-way-to-parse-a-string-to-an-enum
func ParseStringToStateEnum(str string) (State, bool) {
	capabilitiesMap := map[string]State{
		"currency_selection": CurrencySelection,
		"waiting":            Waiting,
		"partially_paid":     PartiallyPaid,
		"paid":               Paid,
		"confirmed":          Confirmed,
		"forwarded":          Forwarded,
		"finished":           Finished,
		"expired":            Expired,
		"failed":             Failed,
	}
	c, ok := capabilitiesMap[strings.ToLower(str)]
	return c, ok
}
