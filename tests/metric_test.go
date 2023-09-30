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
			ExpectedBody:   "Метрика успешно принята: gauge/metricName\n",
		},
		{
			Name:           "Valid JSON",
			RequestBody:    `{"type": "counter", "id": "metricCounter", "delta": 123}`,
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   "Метрика успешно принята: counter/metricCounter\n",
		},
		{
			Name:           "Valid JSON",
			RequestBody:    `{"type": "counter", "id": "PollCount"}`,
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   "Метрика успешно принята: counter/PollCount\n",
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

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			handler := api.NewHandler(storage.NewMetricsStorage())

			req, err := http.NewRequest("POST", "/update/", strings.NewReader(tc.RequestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.UpdateMetricJson(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
			assert.Equal(t, tc.ExpectedBody, w.Body.String())
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
