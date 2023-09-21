package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type metricsStorage interface {
	GetMetricByName(metricName string) (float64, error)
	UpdateMetricValue(metricType string, metricName string, value float64)
	SortMetricByName() []string
	GetAllMetrics() string
}

type Handler struct {
	stor metricsStorage
}

func NewHandler(stor metricsStorage) *Handler {
	return &Handler{
		stor: stor,
	}
}

func (s Handler) GetMetricsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Генерируем HTML страницу
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
