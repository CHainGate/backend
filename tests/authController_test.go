package tests

import (
	"CHainGate/backend/controller"
	"testing"
)

func TestUser(t *testing.T) {
	result := controller.Test()

	if result != "test" {
		t.Errorf("Test() FAILED. Expected %s, got %s", "test", result)
	} else {
		t.Logf("Test() PASSED. Expected %s, got %s", "test", result)
	}
}
