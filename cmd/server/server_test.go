package main

import (
	"errors"
	"github.com/SerjZimmer/devops/internal/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateHandlerBadRequest(t *testing.T) {
	// Create a test HTTP request with an invalid URL
	req, err := http.NewRequest("GET", "/update/invalid/url", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test HTTP ResponseWriter
	rr := httptest.NewRecorder()

	// Call your updateHandler
	api.UpdateHandler(rr, req)

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
			handler := http.HandlerFunc(api.UpdateHandler)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expected, rr.Body.String())
		})
	}
}

func TestValueListHandler(t *testing.T) {
	// Создаем фейковый HTTP-запрос
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	// Создаем фейковый HTTP-ответ
	w := httptest.NewRecorder()

	// Вызываем обработчик
	api.ValueListHandler(w, req)

	// Проверяем, что статус ответа равен http.StatusOK (200)
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что содержимое ответа является HTML
	assert.Contains(t, w.Body.String(), "<html>")
	assert.Contains(t, w.Body.String(), "<h1>Все метрики</h1>")
	assert.Contains(t, w.Body.String(), "</html>")
}

func TestValueHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
		mockStorageErr error
	}{
		{
			name:           "ValidRequest",
			method:         http.MethodGet,
			url:            "/metric/gauge/metric_name",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Неверное имя метрики\n",
			mockStorageErr: nil,
		},
		{
			name:           "InvalidMethod",
			method:         http.MethodPost,
			url:            "/metric/gauge/metric_name",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Метод не разрешен\n",
			mockStorageErr: nil,
		},
		{
			name:           "InvalidURLFormat",
			method:         http.MethodGet,
			url:            "/metric/gauge",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Неверный формат URL\n",
			mockStorageErr: nil,
		},
		{
			name:           "InvalidMetricType",
			method:         http.MethodGet,
			url:            "/metric/invalid_type/metric_name",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Неверный тип метрики\n",
			mockStorageErr: nil,
		},
		{
			name:           "MetricNotFound",
			method:         http.MethodGet,
			url:            "/metric/gauge/unknown_metric",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Неверное имя метрики\n",
			mockStorageErr: errors.New("Metric not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			api.ValueHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

		})
	}
}
