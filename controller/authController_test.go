package controller

import (
	"testing"
)

func TestUser(t *testing.T) {
	result := Test()

	if result != "test" {
		t.Errorf("Test() FAILED. Expected %s, got %s", "test", result)
	} else {
		t.Logf("Test() PASSED. Expected %s, got %s", "test", result)
	}
}
