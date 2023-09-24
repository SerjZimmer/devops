package api

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type metricsStorage interface {
	GetMetricByName(metricName string) (float64, error)                    //возвращать структуру
	UpdateMetricValue(metricType string, metricName string, value float64) // принимать структуру
	SortMetricByName() []string
	GetAllMetrics() string
}

type Handler struct {
	stor   metricsStorage
	logger *zap.Logger
}
type responseWriterWithStatus struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriterWithStatus) Status() int {
	return rw.status
}
func NewHandler(stor metricsStorage) *Handler {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return &Handler{
		stor:   stor,
		logger: logger,
	}
}

//

func (s *Handler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

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

func (s Handler) GetMetricsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, "<html><head><title>Metrics</title></head><body>")
	fmt.Fprintf(w, "<h1>Все метрики</h1>")
	fmt.Fprintf(w, "<ul>")

	fmt.Fprintf(w, "<li> %v </li>", s.stor.GetAllMetrics())

	fmt.Fprintf(w, "</ul></body></html>")
}

func (s Handler) GetMetric(w http.ResponseWriter, r *http.Request) {

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

	value, err := s.stor.GetMetricByName(metricName)
	if err != nil {
		http.Error(w, "Неверное имя метрики", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%v\n", value)

}

func (s Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {

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

	s.stor.UpdateMetricValue(metricType, metricName, value)

	fmt.Fprintf(w, "Метрика успешно принята: %s/%s/%s\n", metricType, metricName, metricValue)
}

func parseNumeric(mValue string) (float64, error) {
	floatVal, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		return 0, err
	}
	return floatVal, nil
}
