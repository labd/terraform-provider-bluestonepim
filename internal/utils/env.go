package utils

import "os"

// Getenv retrieves the value of the environment variable named by the key.
// If the variable is not present in the environment, then the fallback value is returned.
//
// Parameters:
//   - key: The name of the environment variable to retrieve.
//   - fallback: The value to return if the environment variable is not set.
//
// Returns:
//
//	The value of the environment variable if it is set, otherwise the fallback value.
func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
