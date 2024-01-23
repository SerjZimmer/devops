package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/SerjZimmer/devops/internal/storage"
	_ "github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

// metricsStorage представляет интерфейс для взаимодействия с хранилищем метрик.
type metricsStorage interface {
	GetMetricByName(m storage.Metrics) (float64, error)
	UpdateMetricValue(m storage.Metrics) error
	UpdateMetricsValue(m []storage.Metrics) error
	SortMetricByName() []string
	GetAllMetrics() string
	PingDB() error
}

// HashSHA256Middleware представляет middleware для проверки хеша SHA256.
func (s *Handler) HashSHA256Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("HashSHA256")
		if key != "" {
			data := r.URL.Path
			responseHash := calculateHash(data)
			w.Header().Set("HashSHA256", responseHash)
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware представляет middleware для логирования HTTP-запросов и ответов.
func (s *Handler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		text, _ := httputil.DumpRequest(r, true)
		s.logger.Info(string(text))
		logger := s.logger.With(
			zap.String("URI", r.RequestURI),
			zap.String("Method", r.Method),
		)

		rw := &responseWriterWithStatus{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		logger.Info("Request processed",
			zap.Int("Status", rw.status),
			zap.Duration("Duration", time.Since(startTime)),
		)
	})
}

// PingDB обрабатывает HTTP GET-запрос для проверки доступности базы данных.
func (s *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := s.stor.PingDB(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}

// GetMetricsList обрабатывает HTTP GET-запрос для получения списка всех метрик в виде HTML.
func (s *Handler) GetMetricsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	metricsString := s.stor.GetAllMetrics()

	metrics := strings.Split(metricsString, "\n")

	data := struct {
		Metrics []string
	}{
		Metrics: metrics,
	}

	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetMetric обрабатывает HTTP GET-запрос для получения значения конкретной метрики в формате JSON.
func (s *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Неверный формат URL", http.StatusBadRequest)
		return
	}

	metricType := parts[2]
	metricName := parts[3]

	if metricType != "gauge" && metricType != "counter" {
		http.Error(w, "Неверный тип метрики", http.StatusNotFound)
		return
	}

	var m storage.Metrics
	m.ID = metricName

	value, err := s.stor.GetMetricByName(m)
	if err != nil {
		http.Error(w, "Неверное имя метрики", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetMetricJSON обрабатывает HTTP POST-запрос для получения значения метрики из тела запроса в формате JSON.
func (s *Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m storage.Metrics
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		http.Error(w, "Ошибка при разборе JSON", http.StatusBadRequest)
		return
	}

	if m.MType != "gauge" && m.MType != "counter" {
		http.Error(w, "Неверный тип метрики", http.StatusBadRequest)
		return
	}

	value, err := s.stor.GetMetricByName(m)
	if err != nil {
		http.Error(w, "Неверное  имя метрики", http.StatusNotFound)
		return
	}

	if m.MType == "counter" {
		iv := int64(value)
		m.Delta = &iv
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}
	m.Value = &value

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// UpdateMetric обрабатывает HTTP POST-запрос для обновления значения конкретной метрики.
func (s *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Неверный формат URL", http.StatusBadRequest)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	if metricType != "gauge" && metricType != "counter" {
		http.Error(w, "Неверный тип метрики", http.StatusBadRequest)
		return
	}

	value, err := parseNumeric(metricValue)
	if err != nil {
		http.Error(w, "Значение метрики должно быть числом", http.StatusBadRequest)
		return
	}

	var m storage.Metrics
	m.ID = metricName
	m.MType = metricType
	iv := int64(value)
	m.Delta = &iv
	m.Value = &value

	err = s.stor.UpdateMetricValue(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Метрика успешно принята: %s/%s/%s\n", metricType, metricName, metricValue)

}

// UpdateMetricJSON обрабатывает HTTP POST-запрос для обновления значения метрики из тела запроса в формате JSON.
func (s *Handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m storage.Metrics
	buf := bytes.NewBuffer(make([]byte, 0, 1028))

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "Ошибка при разборе r.Body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	decoder := json.NewDecoder(buf)
	if err := decoder.Decode(&m); err != nil {
		http.Error(w, "Ошибка при разборе JSON", http.StatusBadRequest)
		return
	}

	if !isValidMetrics(m) {
		http.Error(w, "Некорректные данные в JSON", http.StatusBadRequest)
		return
	}
	err = s.stor.UpdateMetricValue(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(m)
	if err != nil {
		http.Error(w, "Ошибка при сериализации JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

// UpdateMetricsJSON обрабатывает HTTP POST-запрос для обновления значений множества метрик из тела запроса в формате JSON.
func (s *Handler) UpdateMetricsJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m []storage.Metrics
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		http.Error(w, "Ошибка при разборе JSON", http.StatusBadRequest)
		return
	}

	err := s.stor.UpdateMetricsValue(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(m)
	if err != nil {
		http.Error(w, "Ошибка при сериализации JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

//
