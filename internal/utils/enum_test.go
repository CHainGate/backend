package utils

import "testing"

type ModeEnumString struct {
	enum   Mode
	name   string
	string string
}

type ApiKeyTypeEnumString struct {
	enum   ApiKeyType
	name   string
	string string
}

var modeEnumTests = []ModeEnumString{
	{
		enum:   Main,
		name:   "Main",
		string: "main",
	},
	{
		enum:   Test,
		name:   "Test",
		string: "test",
	},
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

func TestMode_String(t *testing.T) {
	for _, test := range modeEnumTests {
		output := test.enum.String()
		if output != test.string {
			t.Errorf("Expected string of enum %s, to be %s, but got %s", test.name, test.string, output)
		}
	}
}

func TestParseStringToModeEnum(t *testing.T) {
	for _, test := range modeEnumTests {
		output, ok := ParseStringToModeEnum(test.string)
		if output != test.enum {
			t.Errorf("Expected string %s, to be parsed as enum %s, but got %s", test.string, test.name, output)
		}
		if !ok {
			t.Errorf("An error happend in ParseStringToModeEnum!")
		}
	}
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
