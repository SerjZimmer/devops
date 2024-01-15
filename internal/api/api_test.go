package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SerjZimmer/devops/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidMetrics(t *testing.T) {
	// Тест случая, когда метрика является корректной гаугой
	validGauge := storage.Metrics{
		ID:    "ValidGauge",
		MType: "gauge",
		Value: float64Ptr(10.0),
	}
	assert.True(t, isValidMetrics(validGauge), "Expected isValidMetrics to return true for valid gauge metric, but got false")

	// Тест случая, когда метрика является корректным счетчиком
	validCounter := storage.Metrics{
		ID:    "ValidCounter",
		MType: "counter",
		Delta: int64Ptr(1),
	}
	assert.True(t, isValidMetrics(validCounter), "Expected isValidMetrics to return true for valid counter metric, but got false")

	// Тест случая, когда метрика является счетчиком "PollCount"
	pollCount := storage.Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: int64Ptr(1),
	}
	assert.True(t, isValidMetrics(pollCount), "Expected isValidMetrics to return true for PollCount metric, but got false")

	// Тест случая, когда ID метрики отсутствует
	invalidID := storage.Metrics{
		ID:    "",
		MType: "gauge",
		Value: float64Ptr(10.0),
	}
	assert.False(t, isValidMetrics(invalidID), "Expected isValidMetrics to return false for metric with empty ID, but got true")

	// Тест случая, когда MType не равно "gauge" или "counter"
	invalidMType := storage.Metrics{
		ID:    "InvalidMType",
		MType: "invalidType",
		Value: float64Ptr(10.0),
	}
	assert.False(t, isValidMetrics(invalidMType), "Expected isValidMetrics to return false for metric with invalid MType, but got true")

	// Тест случая, когда MType равно "gauge", но значение Value отсутствует
	missingValue := storage.Metrics{
		ID:    "MissingValue",
		MType: "gauge",
		Value: nil,
	}
	assert.False(t, isValidMetrics(missingValue), "Expected isValidMetrics to return false for gauge metric with missing Value, but got true")
}

func int64Ptr(value int) *int64 {
	v := int64(value)
	return &v
}

// Вспомогательная функция для создания указателя на float64
func float64Ptr(value float64) *float64 {
	return &value
}
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

func TestHashSHA256Middleware(t *testing.T) {
	// Создаем хендлер, который будет вызван middleware.
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем middleware с моком хендлера.
	handler := &Handler{}
	middleware := handler.HashSHA256Middleware(mockHandler)

	// Создаем фейковый запрос.
	req, err := http.NewRequest("GET", "/example", nil)
	assert.NoError(t, err)

	// Фейковый ключ для хеша SHA256.
	fakeKey := "fake_key"

	// Добавляем фейковый ключ в заголовок.
	req.Header.Set("HashSHA256", fakeKey)

	// Создаем фейковый ответ.
	w := httptest.NewRecorder()

	// Вызываем middleware.
	middleware.ServeHTTP(w, req)

	// Проверяем, что хеш SHA256 был установлен в заголовке ответа.
	assert.Equal(t, w.Header().Get("HashSHA256"), calculateHash("/example"))
}
