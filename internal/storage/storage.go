package storage

import (
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
)

type MetricsStorage struct {
	Mu         sync.RWMutex
	MetricsMap map[string]float64
	Metrics    Metrics
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		MetricsMap: make(map[string]float64),
		Metrics:    Metrics{},
	}
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (s *MetricsStorage) WriteMetrics(m runtime.MemStats) {
	s.Mu.Lock()
	s.MetricsMap["Alloc"] = float64(m.Alloc)
	s.MetricsMap["BuckHashSys"] = float64(m.BuckHashSys)
	s.MetricsMap["Frees"] = float64(m.Frees)
	s.MetricsMap["GCCPUFraction"] = m.GCCPUFraction
	s.MetricsMap["GCSys"] = float64(m.GCSys)
	s.MetricsMap["HeapAlloc"] = float64(m.HeapAlloc)
	s.MetricsMap["HeapIdle"] = float64(m.HeapIdle)
	s.MetricsMap["HeapInuse"] = float64(m.HeapInuse)
	s.MetricsMap["HeapObjects"] = float64(m.HeapObjects)
	s.MetricsMap["HeapReleased"] = float64(m.HeapReleased)
	s.MetricsMap["HeapSys"] = float64(m.HeapSys)
	s.MetricsMap["LastGC"] = float64(m.LastGC)
	s.MetricsMap["Lookups"] = float64(m.Lookups)
	s.MetricsMap["MCacheInuse"] = float64(m.MCacheInuse)
	s.MetricsMap["MCacheSys"] = float64(m.MCacheSys)
	s.MetricsMap["MSpanInuse"] = float64(m.MSpanInuse)
	s.MetricsMap["MSpanSys"] = float64(m.MSpanSys)
	s.MetricsMap["Mallocs"] = float64(m.Mallocs)
	s.MetricsMap["NextGC"] = float64(m.NextGC)
	s.MetricsMap["NumForcedGC"] = float64(m.NumForcedGC)
	s.MetricsMap["NumGC"] = float64(m.NumGC)
	s.MetricsMap["OtherSys"] = float64(m.OtherSys)
	s.MetricsMap["PauseTotalNs"] = float64(m.PauseTotalNs)
	s.MetricsMap["StackInuse"] = float64(m.StackInuse)
	s.MetricsMap["StackSys"] = float64(m.StackSys)
	s.MetricsMap["Sys"] = float64(m.Sys)
	s.MetricsMap["TotalAlloc"] = float64(m.TotalAlloc)
	s.MetricsMap["PollCount"] = 1
	s.MetricsMap["RandomValue"] = rand.Float64()
	s.Mu.Unlock()
}

func (s *MetricsStorage) UpdateMetricValue(metricType string, metricName string, value float64) {

	s.Mu.Lock()
	if metricType == "counter" {
		if metricName == "PollCount" {
			s.MetricsMap[metricName] += 1
		} else {
			s.MetricsMap[metricName] += value
		}
	} else {
		s.MetricsMap[metricName] = value
	}
	s.Mu.Unlock()
}
func (s *MetricsStorage) GetMetricByName(metricName string) (float64, error) {
	s.Mu.RLock()
	value, exists := s.MetricsMap[metricName]
	s.Mu.RUnlock()
	if exists {
		return value, nil
	}
	return 0, fmt.Errorf("undefind metricName: %v", metricName)
}

func (s *MetricsStorage) SortMetricByName() []string {
	var keys []string
	s.Mu.RLock()
	for key := range s.MetricsMap {
		keys = append(keys, key)
	}
	s.Mu.RUnlock()
	sort.Strings(keys)
	return keys
}

func (s *MetricsStorage) GetAllMetrics() string {
	keys := s.SortMetricByName()
	var result string
	for _, key := range keys {
		result += fmt.Sprintf("%v/%v\n", key, s.MetricsMap[key])
	}
	return result
}
