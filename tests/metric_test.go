package tests

import (
	"errors"
	"github.com/SerjZimmer/devops/internal/api"
	"github.com/SerjZimmer/devops/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_UpdateMetric(t *testing.T) {
	handler := api.NewHandler(storage.NewMetricsStorage())
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid gauge input", `{"id": "metric1", "type": "gauge", "value": 3.14}`, "Метрика успешно принята: gauge/metric1/3.14\n"},
		{"valid counter input", `{"id": "metric2", "type": "counter", "delta": 2}`, "Метрика успешно принята: counter/metric2/2\n"},
		{"invalid JSON", `{"id": "metric3", "type": "gauge", "value": "invalid"}`, "Ошибка при разборе JSON\n"},
		{"invalid metric type", `{"id": "metric4", "type": "invalid", "value": 1.23}`, "Неверный тип метрики\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем POST-запрос с JSON-телом
			reqBody := strings.NewReader(tc.input)
			req, err := http.NewRequest("POST", "/update", reqBody)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h := http.HandlerFunc(handler.UpdateMetric)

			h.ServeHTTP(rr, req)

			assert.Equal(t, tc.expected, rr.Body.String())
		})
	}
}

func Test_GetMetricsList(t *testing.T) {
	handler := api.NewHandler(storage.NewMetricsStorage())

	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	handler.GetMetricsList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), "<html>")
	assert.Contains(t, w.Body.String(), "<h1>Все метрики</h1>")
	assert.Contains(t, w.Body.String(), "</html>")
}

func Test_GetMetric(t *testing.T) {
	handler := api.NewHandler(storage.NewMetricsStorage())

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

			handler.GetMetric(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

		})
	}
}
