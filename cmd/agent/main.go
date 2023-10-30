package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	config "github.com/SerjZimmer/devops/internal/config/agent"
	"github.com/SerjZimmer/devops/internal/storage"
	"net/http"
	"runtime"
	"time"
)

// test
func main() {

	c := config.New()
	s := storage.NewMetricsStorage(c.Storage)
	go func() {
		for {
			poll(s)
			time.Sleep(time.Second * time.Duration(c.PollInterval))
		}
	}()

	for {
		send(s, c)
		sendAllInBatches(s, c, 5)
		time.Sleep(time.Duration(c.ReportInterval) * time.Second)
	}

}

func poll(s *storage.MetricsStorage) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.WriteMetrics(m)
}

func send(s *storage.MetricsStorage, c *config.Config) {

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

		sendMetric(m, c)
	}
	s.Mu.Unlock()
}

func sendAllInBatches(s *storage.MetricsStorage, c *config.Config, batchSize int) {
	s.Mu.Lock()
	var metrics []storage.Metrics

	for metricName, metricValue := range s.MetricsMap {
		m := storage.Metrics{
			ID: metricName,
		}

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
		metrics = append(metrics, m)

		if len(metrics) == batchSize {

			sendMetricsBatch(metrics, c)
			metrics = nil
		}
	}

	if len(metrics) > 0 {
		sendMetricsBatch(metrics, c)
	}

	s.Mu.Unlock()
}

func doReq(data []byte, contentType, path string, c *config.Config) {
	// Create a buffer to store compressed data
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)

	// Write data to a compressed buffer
	_, err := gzipWriter.Write(data)
	if err != nil {
		fmt.Println("Ошибка при сжатии данных:", err)
		return
	}

	// Complete recording and close the compressed buffer
	gzipWriter.Close()

	serverURL := fmt.Sprintf("http://%v/%v/", c.Address, path)

	req, err := http.NewRequest("POST", serverURL, &compressedData)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")
	if c.Key != "" {
		hasher := sha256.New()
		hasher.Write([]byte(c.Key))
		hash := hex.EncodeToString(hasher.Sum(nil))
		req.Header.Set("HashSHA256", hash)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке данных на сервер:", err, serverURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Ошибка при отправке данных на сервер. Код ответа:", resp.StatusCode)
		return
	}
}

func sendMetric(m storage.Metrics, c *config.Config) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Ошибка при маршалинге JSON:", err)
		return
	}

	doReq(jsonData, "application/json", "update", c)
}

func sendMetricsBatch(m []storage.Metrics, c *config.Config) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Ошибка при маршалинге JSON:", err)

	}
	doReq(jsonData, "application/json", "updates", c)
}
