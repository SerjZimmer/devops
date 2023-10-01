package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SerjZimmer/devops/internal/config"
	"github.com/SerjZimmer/devops/internal/storage"
	"net/http"
	"runtime"
	"time"
)

func main() {
	s := storage.NewMetricsStorage()
	c := config.NewConfig()
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(c.PollInterval))
			poll(s, c.Address)
		}
	}()

	for {
		time.Sleep(time.Duration(c.ReportInterval) * time.Second)
		send(s, c.Address)
	}

}

func poll(s *storage.MetricsStorage, address string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.WriteMetrics(m)
}

func send(s *storage.MetricsStorage, address string) {
	s.Mu.Lock()
	var m storage.Metrics
	for metricName, metricValue := range s.MetricsMap {
		m.ID = metricName

		if m.ID != "PollCount" {
			m.MType = "gauge"
			m.Value = &metricValue
		} else {
			m.MType = "counter"
			if m.ID == "PollCount" {
				v := int64(1)
				m.Delta = &v
			} else {
				delta := int64(metricValue)
				m.Delta = &delta
			}
		}

		sendMetric(m, address)
	}
	s.Mu.Unlock()
}

func sendMetric(m storage.Metrics, address string) {

	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Ошибка при маршалинге JSON:", err)
		return
	}

	serverURL := fmt.Sprintf("http://%v/update/", address)

	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке метрики на сервер:", err, serverURL, m)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Ошибка при отправке метрики на сервер. Код ответа:", resp.StatusCode)
		return
	}
}
