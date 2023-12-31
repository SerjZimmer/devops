package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/SerjZimmer/devops/internal/storage"
	_ "github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"net/http/httputil"
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
	GetMetricByName(m storage.Metrics) (float64, error)
	UpdateMetricValue(m storage.Metrics) error
	UpdateMetricsValue(m []storage.Metrics) error
	SortMetricByName() []string
	GetAllMetrics() string
	PingDB() error
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
func (s *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := s.stor.PingDB(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}

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
func calculateHash(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash
}

func (s *Handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m storage.Metrics
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		http.Error(w, "Ошибка при разборе JSON", http.StatusBadRequest)
		return
	}

	if !isValidMetrics(m) {
		http.Error(w, "Некорректные данные в JSON", http.StatusBadRequest)
		return
	}
	err := s.stor.UpdateMetricValue(m)
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

func isValidMetrics(m storage.Metrics) bool {
	if m.ID == "" {
		return false
	}

	if m.MType != "gauge" && m.MType != "counter" {
		return false
	}

	if m.MType == "gauge" && m.Value == nil {
		return false
	}

	if m.MType == "counter" && m.Delta == nil && m.ID != "PollCount" {
		return false
	}

	return true
}

func parseNumeric(mValue string) (float64, error) {
	floatVal, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		return 0, err
	}
	return floatVal, nil
}
