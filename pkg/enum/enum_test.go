package enum

import "testing"

type StateEnumString struct {
	enum   State
	name   string
	string string
}

var stateEnumTests = []StateEnumString{
	{
		enum:   StateWaiting,
		name:   "StateWaiting",
		string: "waiting",
	},
	{
		enum:   StatePartiallyPaid,
		name:   "StatePartiallyPaid",
		string: "partially_paid",
	},
	{
		enum:   StatePaid,
		name:   "StatePaid",
		string: "paid",
	},
	{
		enum:   StateSending,
		name:   "StateSending",
		string: "sending",
	},
	{
		enum:   StateFinished,
		name:   "StateFinished",
		string: "finished",
	},
	{
		enum:   StateExpired,
		name:   "StateExpired",
		string: "expired",
	},
	{
		enum:   StateFailed,
		name:   "StateFailed",
		string: "failed",
	},
}

func TestString(t *testing.T) {
	for _, test := range stateEnumTests {
		output := test.enum.String()
		if output != test.string {
			t.Errorf("Expected string of enum %s, to be %s, but got %s", test.name, test.string, output)
		}
	}
}

func TestParseStringToStateEnum(t *testing.T) {
	for _, test := range stateEnumTests {
		output, ok := ParseStringToStateEnum(test.string)
		if output != test.enum {
			t.Errorf("Expected string %s, to be parsed as enum %s, but got %s", test.string, test.name, output)
		}
		if !ok {
			t.Errorf("An error happend in ParseStringToStateEnum!")
		}
	}
}
