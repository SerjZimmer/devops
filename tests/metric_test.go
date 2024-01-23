package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SerjZimmer/devops/internal/api"
	"github.com/SerjZimmer/devops/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricJson(t *testing.T) {

	testCases := []struct {
		Name           string
		RequestBody    string
		ExpectedStatus int
		ExpectedBody   string
	}{
		{
			Name:           "Valid JSON",
			RequestBody:    `{"type": "gauge", "id": "metricName", "value": 123.45}`,
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   "{\"id\":\"metricName\",\"type\":\"gauge\",\"value\":123.45}",
		},
		{
			Name:           "Valid JSON",
			RequestBody:    `{"type": "counter", "id": "metricCounter", "delta": 123}`,
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   "{\"id\":\"metricCounter\",\"type\":\"counter\",\"delta\":123}",
		},
		{
			Name:           "Valid JSON",
			RequestBody:    `{"type": "counter", "id": "PollCount"}`,
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   "{\"id\":\"PollCount\",\"type\":\"counter\"}",
		},
		{
			Name:           "Invalid JSON",
			RequestBody:    `{"type": "invalid"}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   "Некорректные данные в JSON\n",
		},
		{
			Name:           "Invalid Metric Data",
			RequestBody:    `{"type": "gauge", "id": "", "value": 123.45}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   "Некорректные данные в JSON\n",
		},
	}
	handler := api.NewHandler(storage.TestMetricStorage())
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			req, err := http.NewRequest("POST", "/update/", strings.NewReader(tc.RequestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.UpdateMetricJSON(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
			assert.Equal(t, tc.ExpectedBody, w.Body.String())
		})
	}
}

func Test_GetMetricsList(t *testing.T) {
	handler := api.NewHandler(storage.TestMetricStorage())

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
	handler := api.NewHandler(storage.TestMetricStorage())

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

func TestGetMetricJSON(t *testing.T) {
	handler := api.NewHandler(storage.TestMetricStorage())

	testCases := []struct {
		Name           string
		RequestBody    string
		ExpectedStatus int
		ExpectedBody   string
	}{
		{
			Name:           "Invalid Metric Type",
			RequestBody:    `{"type": "invalid", "id": "metricName"}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   "Неверный тип метрики\n",
		},
		{
			Name:           "Metric Not Found",
			RequestBody:    `{"type": "gauge", "id": "unknown_metric"}`,
			ExpectedStatus: http.StatusNotFound,
			ExpectedBody:   "Неверное  имя метрики\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/get/", strings.NewReader(tc.RequestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.GetMetricJSON(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
			assert.Equal(t, tc.ExpectedBody, w.Body.String())
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	handler := api.NewHandler(storage.TestMetricStorage())

	testCases := []struct {
		Name           string
		URL            string
		ExpectedStatus int
		ExpectedBody   string
	}{
		{
			Name:           "Valid Request",
			URL:            "/metric/gauge/metricName/123.45",
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   "Метрика успешно принята: gauge/metricName/123.45\n",
		},

		{
			Name:           "Invalid URL Format",
			URL:            "/metric/gauge/metricName",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   "Неверный формат URL\n",
		},
		{
			Name:           "Invalid Metric Type",
			URL:            "/metric/invalid/metricName/123.45",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   "Неверный тип метрики\n",
		},
		{
			Name:           "Invalid Metric Value",
			URL:            "/metric/gauge/metricName/invalid_value",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   "Значение метрики должно быть числом\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.URL, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.UpdateMetric(rr, req)

			assert.Equal(t, tc.ExpectedStatus, rr.Code)
			assert.Equal(t, tc.ExpectedBody, rr.Body.String())
		})
	}
}

func TestUpdateMetricsJSON(t *testing.T) {
	handler := api.NewHandler(storage.TestMetricStorage())

	// Подготовка тестовых данных
	testMetrics := []storage.Metrics{
		{ID: "metric1", MType: "gauge", Value: float64Ptr(10.5)},
		{ID: "metric2", MType: "counter", Delta: int64Ptr(5)},
		// Добавьте нужные метрики для тестирования
	}

	// Преобразование тестовых данных в JSON
	jsonData, err := json.Marshal(testMetrics)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/update-metrics/", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.UpdateMetricsJSON(w, req)

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверка тела ответа
	var responseMetrics []storage.Metrics
	err = json.Unmarshal(w.Body.Bytes(), &responseMetrics)
	assert.NoError(t, err)

	// Добавьте проверки для ожидаемых значений в ответе
	assert.ElementsMatch(t, testMetrics, responseMetrics)
}

// Вспомогательная функция для создания указателя на float64
func float64Ptr(value float64) *float64 {
	return &value
}

// Вспомогательная функция для создания указателя на int64
func int64Ptr(value int64) *int64 {
	return &value
}
