package main

import (
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
	monitoring(s, c.Address, c.PollInterval, c.ReportInterval)
}

func monitoring(s *storage.MetricsStorage, address string, pollInterval, reportInterval int) {

	for {
		go func() {
			for {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				s.WriteMetrics(m)
				time.Sleep(time.Duration(pollInterval) * time.Second)
			}
		}()

		go func() {
			for {
				s.Mu.Lock()
				for metricName, metricValue := range s.MetricsMap {

					s.Metrics.ID = metricName

					if s.Metrics.ID != "PollCount" {
						s.Metrics.MType = "gauge"
						s.Metrics.Value = &metricValue
						go sendMetric(s, address)
					} else {
						s.Metrics.MType = "counter"
						delta := int64(metricValue)
						s.Metrics.Delta = &delta
						go sendMetric(s, address)
					}
				}
				s.Mu.Unlock()

				time.Sleep(time.Duration(reportInterval) * time.Second)
			}
		}()
	}
}

func sendMetric(s *storage.MetricsStorage, address string) {

	serverURL := fmt.Sprintf("http://%v/update/%v/%v/%v", address, s.Metrics.MType, s.Metrics, s.Metrics.Value)
	if s.Metrics.MType == "counter" {
		serverURL = fmt.Sprintf("http://%v/update/%v/%v/%v", address, s.Metrics.MType, s.Metrics, s.Metrics.Delta)
	}

	req, err := http.NewRequest("POST", serverURL, nil)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		//fmt.Println("Ошибка при отправке метрики на сервер:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Ошибка при отправке метрики на сервер. Код ответа:", resp.StatusCode)
		return
	}
}
