package main

import (
	"github.com/SerjZimmer/devops/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock sendMetric function for testing
var (
	originalSendMetric = sendMetric
	sendMetricT        = func(s *storage.MetricsStorage, address string) {}
)
var strg = storage.NewMetricsStorage()

func Test_sendMetric(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with the original one after the test
	defer func() { sendMetricT = originalSendMetric }()

	// Test sendMetric with test data
	strg.Metrics.MType = "gauge"
	strg.Metrics.ID = "metricName"
	v := 123.45
	strg.Metrics.Value = &v

	// Call sendMetric
	sendMetric(strg, "localhost:8080")
}

func Test_sendMetricCounter(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with the original one after the test
	defer func() { sendMetricT = originalSendMetric }()

	// Test sendMetric with test data
	strg.Metrics.MType = "counter"
	strg.Metrics.ID = "metricName"
	v := 123
	vi := int64(v)
	strg.Metrics.Delta = &vi

	// Call sendMetric
	sendMetric(strg, "localhost:8080")
}

func Test_monitoring(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with a mock for testing
	sendMetricT = func(s *storage.MetricsStorage, address string) {
		// Mock behavior for sendMetric function during testing
		// You can add assertions or checks here as needed
		assert.Equal(t, "gauge", strg.Metrics.MType)
		assert.Equal(t, "metricName", strg.Metrics.ID)
		// Check metricValue against expected values
	}
	go monitoring(strg, "localhost:8080", 10, 10)
}
