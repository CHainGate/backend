package enum

import "strings"

type State int

// https://levelup.gitconnected.com/implementing-enums-in-golang-9537c433d6e2
const (
	Waiting State = iota + 1
	PartiallyPaid
	Paid
	Sending
	Finished
	Expired
	Failed
)

func (s State) String() string {
	return [...]string{"waiting", "partially_paid", "paid", "sending", "finished", "expired", "failed"}[s-1]
}

// ParseStringToStateEnum ParseStringToModeEnum https://stackoverflow.com/questions/68543604/best-way-to-parse-a-string-to-an-enum
func ParseStringToStateEnum(str string) (State, bool) {
	capabilitiesMap := map[string]State{
		"waiting":        Waiting,
		"partially_paid": PartiallyPaid,
		"paid":           Paid,
		"sending":        Sending,
		"finished":       Finished,
		"expired":        Expired,
		"failed":         Failed,
	}
	c, ok := capabilitiesMap[strings.ToLower(str)]
	return c, ok
}
