package utils

import "strings"

type Mode int
type ApiKeyType int

// https://levelup.gitconnected.com/implementing-enums-in-golang-9537c433d6e2
const (
	Main   Mode       = 1
	Test   Mode       = 2
	Public ApiKeyType = 1
	Secret ApiKeyType = 2
)

func (m Mode) String() string {
	return [...]string{"main", "test"}[m-1]
}

func (a ApiKeyType) String() string {
	return [...]string{"public", "secret"}[a-1]
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

func ParseStringToApiKeyTypeEnum(str string) (ApiKeyType, bool) {
	capabilitiesMap := map[string]ApiKeyType{
		"public": Public,
		"secret": Secret,
	}
	c, ok := capabilitiesMap[strings.ToLower(str)]
	return c, ok
}
