package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	config "github.com/SerjZimmer/devops/internal/config/agent"
	"github.com/SerjZimmer/devops/internal/storage"
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

func TestCompressData(t *testing.T) {
	// Тест для функции compressData

	// Подготовка данных
	data := []byte("test data")

	// Вызов функции compressData
	compressedData, err := compressData(data)
	assert.NoError(t, err)

	// Проверка, что данные не пусты
	assert.NotEmpty(t, compressedData.Bytes())

	// Проверка, что функция не вернула ошибку
	assert.NoError(t, err)

	// Проверка, что данные действительно сжаты
	reader, err := gzip.NewReader(&compressedData)
	assert.NoError(t, err)
	defer reader.Close()

	var decompressedData bytes.Buffer
	_, err = decompressedData.ReadFrom(reader)
	assert.NoError(t, err)

	// Проверка, что исходные данные совпадают с разжатыми данными
	assert.Equal(t, data, decompressedData.Bytes())
}

func TestDoReq(t *testing.T) {
	// Тест для функции doReq

	// Подготовка тестовых данных
	testData := []byte("test data")
	contentType := "application/json"
	path := "example"
	config := &config.Config{
		Address: "localhost:8080",
		Key:     "secretKey",
	}

	// Mock сервера для тестирования HTTP запросов
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверка корректности запроса
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, contentType, r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		// Проверка наличия хэша в запросе, если указан ключ в конфиге
		if config.Key != "" {
			hasher := sha256.New()
			hasher.Write([]byte(config.Key))
			expectedHash := hex.EncodeToString(hasher.Sum(nil))
			assert.Equal(t, expectedHash, r.Header.Get("HashSHA256"))
		}

		// Чтение тела запроса и проверка его содержимого
		reader, err := gzip.NewReader(r.Body)
		assert.NoError(t, err)
		defer reader.Close()

		var decompressedData bytes.Buffer
		_, err = decompressedData.ReadFrom(reader)
		assert.NoError(t, err)

		assert.Equal(t, testData, decompressedData.Bytes())

		// Отправка успешного ответа
		w.WriteHeader(http.StatusOK)
	}))

	// Завершение работы сервера при завершении теста
	defer server.Close()

	// Установка адреса тестового сервера
	config.Address = server.Listener.Addr().String()

	// Вызов функции doReq
	doReq(testData, contentType, path, config)
}
