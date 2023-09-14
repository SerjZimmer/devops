package api

import (
	"fmt"
	"github.com/SerjZimmer/devops/internal/storage"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var mu sync.Mutex

func ValueListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	keys := storage.SortMetricByName()

	// Генерируем HTML страницу
	fmt.Fprintf(w, "<html><head><title>Metrics</title></head><body>")
	fmt.Fprintf(w, "<h1>Все метрики</h1>")
	fmt.Fprintf(w, "<ul>")
	for _, key := range keys {
		fmt.Fprintf(w, "<li>%s: %v</li>", key, storage.MetricsMap[key])
	}
	fmt.Fprintf(w, "</ul></body></html>")
}

func ValueHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	// Разбираем URL-параметры
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

	mu.Lock()
	value, err := storage.CheckMetricByName(metricName)
	if err != nil {
		http.Error(w, "Неверное имя метрики", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%v\n", value)
	mu.Unlock()

}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

	// Разбираем URL-параметры
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

	mu.Lock()
	storage.UpdateMetricValue(metricType, metricName, value)
	mu.Unlock()
	// Возвращаем успешный ответ
	fmt.Fprintf(w, "Метрика успешно принята: %s/%s/%s\n", metricType, metricName, metricValue)
}

func parseNumeric(mValue string) (float64, error) {
	floatVal, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		return 0, err
	}
	return floatVal, nil
}
