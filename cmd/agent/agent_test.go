package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock sendMetric function for testing
var (
	originalSendMetric = sendMetric
	sendMetricT        = func(metricType, metricName string, metricValue any, address string) {}
)

func TestSendMetric(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with the original one after the test
	defer func() { sendMetricT = originalSendMetric }()

	// Test sendMetric with test data
	metricType := "gauge"
	metricName := "metricName"
	metricValue := 123.45

	// Call sendMetric
	sendMetric(metricType, metricName, metricValue, "localhost:8080")
}

func TestSendMetricCounter(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with the original one after the test
	defer func() { sendMetricT = originalSendMetric }()

	// Test sendMetric with test data
	metricType := "counter"
	metricName := "PollCount"
	metricValue := 123.45

	// Call sendMetric
	sendMetric(metricType, metricName, metricValue, "localhost:8080")
}

func TestMonitoring(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with a mock for testing
	sendMetricT = func(metricType, metricName string, metricValue any, address string) {
		// Mock behavior for sendMetric function during testing
		// You can add assertions or checks here as needed
		assert.Equal(t, "gauge", metricType)
		assert.Equal(t, "metricName", metricName)
		// Check metricValue against expected values
	}
	go monitoring("localhost:8080", 10, 10)
}
