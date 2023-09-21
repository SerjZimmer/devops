package api

import (
	"github.com/stretchr/testify/assert"
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
