package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_parseNumeric(t *testing.T) {
	tests := []struct {
		input       string
		expectedVal float64
		expectedErr bool
	}{
		{"123.45", 123.45, false},
		{"-67.89", -67.89, false},
		{"not_a_number", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, err := parseNumeric(tt.input)

			if tt.expectedErr {
				assert.Error(t, err, "Expected an error for input %s", tt.input)
			} else {
				assert.NoError(t, err, "Expected no error for input %s", tt.input)
				assert.Equal(t, tt.expectedVal, val, "Expected value %f for input %s", tt.expectedVal, tt.input)
			}
		})
	}
}

func Test_calculateHash(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"world", "486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := calculateHash(tc.input)
			require.Equal(t, tc.expected, result, "Expected value %f for input %s", tc.expected, result)
		})
	}
}
