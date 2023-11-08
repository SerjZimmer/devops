package main

import (
	config "github.com/SerjZimmer/devops/internal/config/agent"
	"github.com/SerjZimmer/devops/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock sendMetric function for testing
var (
	originalSendMetric = sendMetric
	sendMetricT        = func(m storage.Metrics, c *config.Config) {}
)

var m storage.Metrics

func Test_sendMetric(t *testing.T) {
	c := config.New()
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with the original one after the test
	defer func() { sendMetricT = originalSendMetric }()

	// Test sendMetric with test data
	m.MType = "gauge"
	m.ID = "metricName"
	v := 123.45
	m.Value = &v

	// Call sendMetric
	sendMetric(m, c)
}

func Test_sendMetricCounter(t *testing.T) {
	c := config.Config{}
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Replace the sendMetric function with the original one after the test
	defer func() { sendMetricT = originalSendMetric }()

	// Test sendMetric with test data
	m.MType = "counter"
	m.ID = "metricName"
	v := 123
	vi := int64(v)
	m.Delta = &vi

	// Call sendMetric
	sendMetric(m, &c)
}
