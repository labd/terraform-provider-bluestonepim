package utils

import (
	"testing"
)

func TestGetenvReturnsEnvValue(t *testing.T) {
	t.Setenv("TEST_ENV", "value")

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
	t.Setenv("EMPTY_ENV", "")

	result := GetEnv("EMPTY_ENV", "default")
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}
}
