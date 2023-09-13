package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	address        string
	pollInterval   int
	reportInterval int
	metricsMap     = map[string]float64{}
)

func flagInit() {
	flag.StringVar(&address, "a", getEnv("ADDRESS", "localhost:8080"), "Адрес эндпоинта HTTP-сервера")
	flag.IntVar(&reportInterval, "r", getEnvAsInt("REPORT_INTERVAL", 10), "Частота отправки метрик на сервер")
	flag.IntVar(&pollInterval, "p", getEnvAsInt("POLL_INTERVAL", 2), "Частота опроса метрик из пакета runtime")
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "a" || f.Name == "r" || f.Name == "p" {
			return
		}
		fmt.Printf("Неизвестный флаг: -%s\n", f.Name)
		flag.PrintDefaults()
		os.Exit(1)
	})
	flag.Parse()
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err == nil {
			return value
		}
	}
	return defaultValue
}

func main() {
	flagInit()

	go func() {

		r := mux.NewRouter()

		r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", UpdateHandler).Methods("POST")
		r.HandleFunc("/value/{metricType}/{metricName}", ValueHandler).Methods("GET")
		r.HandleFunc("/", ValueListHandler).Methods("GET")

		http.Handle("/", r)
		if err := run(); err != nil {
			panic(err)
		}

	}()
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
		<-sigchan

		close(shutdownChan) // Отправляем сигнал завершения серверу
	}()
	time.Sleep(time.Second)

	<-shutdownChan
	// Остановка сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при завершении работы сервера: %v\n", err)
	}

	os.Exit(0)
}

var (
	server       *http.Server
	shutdownChan = make(chan struct{})
)

func run() error {
	fmt.Printf("Сервер запущен на %v\n", address)

	server = &http.Server{Addr: address}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-shutdownChan // Ждем сигнала завершения
	fmt.Println("Завершение работы сервера...")

	// Завершаем работу сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при завершении работы сервера: %v\n", err)
	}

	return nil
}

var mu sync.Mutex

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
	if metricType == "counter" {
		metricsMap[metricName] += value
	} else {
		metricsMap[metricName] = value
	}

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
	value, exists := metricsMap[metricName]
	mu.Unlock()
	if exists {
		fmt.Fprintf(w, "%v\n", value)
	} else {
		http.Error(w, "Неверное имя метрики", http.StatusNotFound)
		return
	}

}

func ValueListHandler(w http.ResponseWriter, r *http.Request) {
	// Заголовок ответа, указываем тип контента как text/html
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Извлекаем ключи из мапы MetricsMap
	var keys []string
	for key := range metricsMap {
		keys = append(keys, key)
	}

	// Сортируем ключи
	sort.Strings(keys)

	// Генерируем HTML страницу
	fmt.Fprintf(w, "<html><head><title>Metrics</title></head><body>")
	fmt.Fprintf(w, "<h1>Все метрики</h1>")
	fmt.Fprintf(w, "<ul>")
	for _, key := range keys {
		fmt.Fprintf(w, "<li>%s: %v</li>", key, metricsMap[key])
	}
	fmt.Fprintf(w, "</ul></body></html>")
}
