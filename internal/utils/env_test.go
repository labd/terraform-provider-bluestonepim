package utils

import (
	"os"
	"testing"
)

func TestGetenvReturnsEnvValue(t *testing.T) {
	os.Setenv("TEST_ENV", "value")
	defer os.Unsetenv("TEST_ENV")

	result := GetEnv("TEST_ENV", "default")
	if result != "value" {
		t.Errorf("expected 'value', got '%s'", result)
	}
}

func TestGetenvReturnsFallbackWhenEnvNotSet(t *testing.T) {
	result := GetEnv("NON_EXISTENT_ENV", "default")
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}
}

func TestGetenvReturnsFallbackWhenEnvIsEmpty(t *testing.T) {
	os.Setenv("EMPTY_ENV", "")
	defer os.Unsetenv("EMPTY_ENV")

	result := GetEnv("EMPTY_ENV", "default")
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}
}
