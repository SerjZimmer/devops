package api

import (
	"encoding/json"
	"fmt"
	"github.com/SerjZimmer/devops/internal/storage"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const metricsListTemplate = `
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Все метрики</h1>
    <ul>
        {{range .Metrics}}
        <li>{{.}}</li>
        {{end}}
    </ul>
</body>
</html>
`

var tmpl = template.Must(template.New("metricsList").Parse(metricsListTemplate))

type metricsStorage interface {
	GetMetricByName(metricName string) (float64, error) //возвращать структуру
	UpdateMetricValue(m storage.Metrics)                // принимать структуру
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

	metricsString := s.stor.GetAllMetrics()

	metrics := strings.Split(metricsString, "\n")

	data := struct {
		Metrics []string
	}{
		Metrics: metrics,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	if m.MType != "counter" {
		s.stor.UpdateMetricValue(m)
		fmt.Fprintf(w, "Метрика успешно принята: %s/%s/%v\n", m.MType, m.ID, *m.Value)
	} else {
		s.stor.UpdateMetricValue(m)
		fmt.Fprintf(w, "Метрика успешно принята: %s/%s/%v\n", m.MType, m.ID, *m.Delta)
	}

}

func parseNumeric(mValue string) (float64, error) {
	floatVal, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		return 0, err
	}
	return floatVal, nil
}
