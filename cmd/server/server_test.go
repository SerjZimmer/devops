package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseNumeric(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected float64
	}{
		{"valid input", "3.14", 3.14},
		{"invalid input", "invalid", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := parseNumeric(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestUpdateHandlerBadRequest(t *testing.T) {
	// Create a test HTTP request with an invalid URL
	req, err := http.NewRequest("GET", "/update/invalid/url", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test HTTP ResponseWriter
	rr := httptest.NewRecorder()

	// Call your updateHandler
	UpdateHandler(rr, req)

	// Check if the response status code is http.StatusBadRequest
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Check the response body for the error message
	expectedResponse := "Неверный формат URL\n"
	assert.Equal(t, expectedResponse, rr.Body.String())
}
func TestUpdateHandler(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid gauge input", "/update/gauge/metric1/3.14", "Метрика успешно принята: gauge/metric1/3.14\n"},
		{"valid counter input", "/update/counter/metric2/2.71", "Метрика успешно принята: counter/metric2/2.71\n"},
		{"invalid URL format", "/update/gauge/metric3", "Неверный формат URL\n"},
		{"invalid metric type", "/update/invalid/metric4/1.23", "Неверный тип метрики\n"},
		{"invalid metric value", "/update/gauge/metric5/invalid", "Значение метрики должно быть числом\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.input, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(UpdateHandler)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expected, rr.Body.String())
		})
	}
}
