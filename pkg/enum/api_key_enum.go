package enum

import "strings"

type ApiKeyType int

// https://levelup.gitconnected.com/implementing-enums-in-golang-9537c433d6e2
const (
	Public ApiKeyType = iota + 1
	Secret
)

func (a ApiKeyType) String() string {
	return [...]string{"public", "secret"}[a-1]
}

func ParseStringToApiKeyTypeEnum(str string) (ApiKeyType, bool) {
	capabilitiesMap := map[string]ApiKeyType{
		"public": Public,
		"secret": Secret,
	}
	c, ok := capabilitiesMap[strings.ToLower(str)]
	return c, ok
}
