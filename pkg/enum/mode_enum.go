package enum

import "strings"

type Mode int

// https://levelup.gitconnected.com/implementing-enums-in-golang-9537c433d6e2
const (
	Main Mode = iota + 1
	Test
)

func (m Mode) String() string {
	return [...]string{"main", "test"}[m-1]
}

// ParseStringToModeEnum https://stackoverflow.com/questions/68543604/best-way-to-parse-a-string-to-an-enum
func ParseStringToModeEnum(str string) (Mode, bool) {
	capabilitiesMap := map[string]Mode{
		"main": Main,
		"test": Test,
	}
	c, ok := capabilitiesMap[strings.ToLower(str)]
	return c, ok
}
