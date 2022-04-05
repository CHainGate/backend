package enum

import "testing"

type ApiKeyTypeEnumString struct {
	enum   ApiKeyType
	name   string
	string string
}

var apiKeyTypeEnumTests = []ApiKeyTypeEnumString{
	{
		enum:   Public,
		name:   "Public",
		string: "public",
	},
	{
		enum:   Secret,
		name:   "Secret",
		string: "secret",
	},
}

func TestApiKeyType_String(t *testing.T) {
	for _, test := range apiKeyTypeEnumTests {
		output := test.enum.String()
		if output != test.string {
			t.Errorf("Expected string of enum %s, to be %s, but got %s", test.name, test.string, output)
		}
	}
}

func TestParseStringToApiKeyTypeEnum(t *testing.T) {
	for _, test := range apiKeyTypeEnumTests {
		output, ok := ParseStringToApiKeyTypeEnum(test.string)
		if output != test.enum {
			t.Errorf("Expected string %s, to be parsed as enum %s, but got %s", test.string, test.name, output)
		}
		if !ok {
			t.Errorf("An error happend in ParseStringToApiKeyTypeEnum!")
		}
	}
}
