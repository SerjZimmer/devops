package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/SerjZimmer/devops/internal/config"
	"github.com/SerjZimmer/devops/internal/storage"
	"net/http"
	"runtime"
	"time"
)

func main() {

	c := config.NewConfig()
	s := storage.NewMetricsStorage(c)
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(c.PollInterval))
			poll(s, c.Address)
		}
	}()

	for {
		time.Sleep(time.Duration(c.ReportInterval) * time.Second)
		send(s, c.Address)
		sendAll(s, c.Address)
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

func sendAll(s *storage.MetricsStorage, address string) {
	s.Mu.Lock()
	var metrics []storage.Metrics
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
		metrics = append(metrics, m)

	}
	sendMetrics(metrics, address)
	s.Mu.Unlock()
}

func sendCompressedContent(data []byte, contentType string, address string) {
	// Создание буфера для хранения сжатых данных
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)

	// Запись данных в сжатый буфер
	_, err := gzipWriter.Write(data)
	if err != nil {
		fmt.Println("Ошибка при сжатии данных:", err)
		return
	}

	// Завершение записи и закрытие сжатого буфера
	gzipWriter.Close()

	serverURL := fmt.Sprintf("http://%v/update/", address)

	// Создание HTTP-запроса с сжатыми данными
	req, err := http.NewRequest("POST", serverURL, &compressedData)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	// Установка заголовков для указания сжатого формата и типа контента
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")

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

func sendAllCompressedContent(data []byte, contentType string, address string) {
	// Создание буфера для хранения сжатых данных
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)

	// Запись данных в сжатый буфер
	_, err := gzipWriter.Write(data)
	if err != nil {
		fmt.Println("Ошибка при сжатии данных:", err)
		return
	}

	// Завершение записи и закрытие сжатого буфера
	gzipWriter.Close()

	serverURL := fmt.Sprintf("http://%v/updates/", address)

	// Создание HTTP-запроса с сжатыми данными
	req, err := http.NewRequest("POST", serverURL, &compressedData)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	// Установка заголовков для указания сжатого формата и типа контента
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")

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

func sendMetric(m storage.Metrics, address string) {
	// Маршалинг JSON-данных
	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Ошибка  при маршалинге JSON:", err)
		return
	}

	// Определение типа контента (application/json) и отправка сжатых данных
	sendCompressedContent(jsonData, "application/json", address)
}

func sendMetrics(m []storage.Metrics, address string) {
	// Маршалинг JSON-данных
	jsonData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Ошибка  при маршалинге JSON:", err)
		return
	}

	// Определение типа контента (application/json) и отправка сжатых данных
	sendAllCompressedContent(jsonData, "application/json", address)
}
