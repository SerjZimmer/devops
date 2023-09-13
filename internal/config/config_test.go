package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_getEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		envValue      string
		expectedValue string
	}{
		{
			name:          "KeyExists",
			key:           "EXISTING_KEY",
			defaultValue:  "default_value",
			envValue:      "test_value",
			expectedValue: "test_value",
		},
		{
			name:          "KeyNotExists",
			key:           "NON_EXISTENT_KEY",
			defaultValue:  "default_value",
			envValue:      "",
			expectedValue: "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)

			assert.Equal(t, tt.expectedValue, result, "Expected value mismatch")
		})
	}
}

func Test_getEnvAsInt(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  int
		envValue      string
		expectedValue int
		expectError   bool
	}{
		{
			name:          "KeyExistsValidInt",
			key:           "EXISTING_INT_KEY",
			defaultValue:  42,
			envValue:      "123",
			expectedValue: 123,
			expectError:   false,
		},
		{
			name:          "KeyExistsInvalidInt",
			key:           "EXISTING_INT_KEY",
			defaultValue:  42,
			envValue:      "not_an_int",
			expectedValue: 42,
			expectError:   true,
		},
		{
			name:          "KeyNotExists",
			key:           "NON_EXISTENT_INT_KEY",
			defaultValue:  42,
			envValue:      "",
			expectedValue: 42,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnvAsInt(tt.key, tt.defaultValue)

			assert.Equal(t, tt.expectedValue, result, "Expected value mismatch")
		})
	}
}
