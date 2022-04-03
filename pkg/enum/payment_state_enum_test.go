package enum

import "testing"

type StateEnumString struct {
	enum   State
	name   string
	string string
}

var stateEnumTests = []StateEnumString{
	{
		enum:   Waiting,
		name:   "Waiting",
		string: "waiting",
	},
	{
		enum:   PartiallyPaid,
		name:   "PartiallyPaid",
		string: "partially_paid",
	},
	{
		enum:   Paid,
		name:   "StatePaid",
		string: "paid",
	},
	{
		enum:   Sending,
		name:   "StateSending",
		string: "sending",
	},
	{
		enum:   Finished,
		name:   "Finished",
		string: "finished",
	},
	{
		enum:   Expired,
		name:   "StateExpired",
		string: "expired",
	},
	{
		enum:   Failed,
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
