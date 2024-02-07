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
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// poll собирает текущие метрики использования памяти и записывает их в хранилище.
func poll(s *storage.MetricsStorage) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.WriteMetrics(m)
}

// send отправляет каждую метрику из хранилища на сервер.
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

// sendAllInBatches отправляет метрики на сервер пакетами заданного размера.
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

// doReq выполняет HTTP-запрос на сервер с сжатием данных.
func doReq(data []byte, contentType, path string, c *config.Config) {
	compressedData, err := compressData(data)
	if err != nil {
		fmt.Println("Ошибка при сжатии данных:", err)
		return
	}

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

// compressData сжимает данные с использованием Gzip.
func compressData(data []byte) (bytes.Buffer, error) {
	// Create a buffer to store compressed data
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)

	// Write data to the compressed buffer
	_, err := gzipWriter.Write(data)
	if err != nil {
		return compressedData, fmt.Errorf("ошибка при записи сжатых данных: %v", err)
	}

	// Complete writing and close the compressed buffer
	err = gzipWriter.Close()
	if err != nil {
		return compressedData, fmt.Errorf("ошибка при закрытии сжатого буфера: %v", err)
	}

	return compressedData, nil
}

// sendMetric отправляет отдельную метрику на сервер.
func sendMetric(m storage.Metrics, c *config.Config) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Ошибка при маршалинге JSON:", err)
		return
	}

	doReq(jsonData, "application/json", "update", c)
}

// sendMetricsBatch отправляет пакет метрик на сервер.
func sendMetricsBatch(m []storage.Metrics, c *config.Config) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Ошибка при маршалинге JSON:", err)

	}
	doReq(jsonData, "application/json", "updates", c)
}

func printBuildInfo() {
	fmt.Println("Build version:", getOrDefault(buildVersion, "N/A"))
	fmt.Println("Build date:", getOrDefault(buildDate, "N/A"))
	fmt.Println("Build commit:", getOrDefault(buildCommit, "N/A"))
}

func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
