package main

import (
	"fmt"
	"github.com/SerjZimmer/devops/internal/config"
	"github.com/SerjZimmer/devops/internal/storage"
	"net/http"
	"sync"
	"time"
)

func main() {
	strg := storage.NewMetricsStorage()
	config.FlagInit()
	for {
		monitoring(strg, config.Address, config.PollInterval, config.ReportInterval)
	}
}

var mu sync.Mutex

func monitoring(strg *storage.MetricsStorage, address string, pollInterval, reportInterval int) {

	go func() {
		for {
			strg.WriteMetrics()
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

	go func() {
		for {
			mu.Lock()
			for metricName, metricValue := range strg.MetricsMap {
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

	select {}
}

func sendMetric(metricType, metricName string, metricValue any, address string) {
	serverURL := fmt.Sprintf("http://%v/update/%s/%s/%v", address, metricType, metricName, metricValue)

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
