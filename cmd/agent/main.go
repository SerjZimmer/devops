package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
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
	for {
		monitoring(address, pollInterval, reportInterval)
	}
}

var mu sync.Mutex

func monitoring(address string, pollInterval, reportInterval int) {
	// Горутина для сбора метрик
	go func() {
		for {
			// Собираем метрики о памяти
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// Обновляем метрики из пакета runtime
			mu.Lock()
			metricsMap["Alloc"] = float64(m.Alloc)
			metricsMap["BuckHashSys"] = float64(m.BuckHashSys)
			metricsMap["Frees"] = float64(m.Frees)
			metricsMap["GCCPUFraction"] = m.GCCPUFraction
			metricsMap["GCSys"] = float64(m.GCSys)
			metricsMap["HeapAlloc"] = float64(m.HeapAlloc)
			metricsMap["HeapIdle"] = float64(m.HeapIdle)
			metricsMap["HeapInuse"] = float64(m.HeapInuse)
			metricsMap["HeapObjects"] = float64(m.HeapObjects)
			metricsMap["HeapReleased"] = float64(m.HeapReleased)
			metricsMap["HeapSys"] = float64(m.HeapSys)
			metricsMap["LastGC"] = float64(m.LastGC)
			metricsMap["Lookups"] = float64(m.Lookups)
			metricsMap["MCacheInuse"] = float64(m.MCacheInuse)
			metricsMap["MCacheSys"] = float64(m.MCacheSys)
			metricsMap["MSpanInuse"] = float64(m.MSpanInuse)
			metricsMap["MSpanSys"] = float64(m.MSpanSys)
			metricsMap["Mallocs"] = float64(m.Mallocs)
			metricsMap["NextGC"] = float64(m.NextGC)
			metricsMap["NumForcedGC"] = float64(m.NumForcedGC)
			metricsMap["NumGC"] = float64(m.NumGC)
			metricsMap["OtherSys"] = float64(m.OtherSys)
			metricsMap["PauseTotalNs"] = float64(m.PauseTotalNs)
			metricsMap["StackInuse"] = float64(m.StackInuse)
			metricsMap["StackSys"] = float64(m.StackSys)
			metricsMap["Sys"] = float64(m.Sys)
			metricsMap["TotalAlloc"] = float64(m.TotalAlloc)
			metricsMap["PollCount"] = metricsMap["PollCount"] + 1
			metricsMap["RandomValue"] = rand.Float64()
			mu.Unlock()

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

	// Горутина для отправки метрик на сервер
	go func() {
		for {

			mu.Lock()
			for metricName, metricValue := range metricsMap {
				if metricName != "PollCount" {
					go sendMetric("gauge", metricName, metricValue, address)
				} else {
					go sendMetric("counter", metricName, int64(metricValue), address)
				}
			}
			mu.Unlock()

			time.Sleep(time.Duration(reportInterval) * time.Second)
		}
	}()

	// Ожидание завершения работы агента
	select {}
}

// Отправка метрики на сервер
func sendMetric(metricType, metricName string, metricValue any, address string) {
	serverURL := fmt.Sprintf("http://%v/update/%s/%s/%v", address, metricType, metricName, metricValue)

	// Отправляем POST-запрос на сервер
	req, err := http.NewRequest("POST", serverURL, nil)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке метрики на сервер:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Ошибка при отправке метрики на сервер. Код ответа:", resp.StatusCode)
		return
	}
}
