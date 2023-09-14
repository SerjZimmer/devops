package storage

import (
	"fmt"
	"math/rand"
	"runtime"
	"sort"
)

var MetricsMap = map[string]float64{}

func WriteMetrics(m runtime.MemStats) {
	MetricsMap["Alloc"] = float64(m.Alloc)
	MetricsMap["BuckHashSys"] = float64(m.BuckHashSys)
	MetricsMap["Frees"] = float64(m.Frees)
	MetricsMap["GCCPUFraction"] = m.GCCPUFraction
	MetricsMap["GCSys"] = float64(m.GCSys)
	MetricsMap["HeapAlloc"] = float64(m.HeapAlloc)
	MetricsMap["HeapIdle"] = float64(m.HeapIdle)
	MetricsMap["HeapInuse"] = float64(m.HeapInuse)
	MetricsMap["HeapObjects"] = float64(m.HeapObjects)
	MetricsMap["HeapReleased"] = float64(m.HeapReleased)
	MetricsMap["HeapSys"] = float64(m.HeapSys)
	MetricsMap["LastGC"] = float64(m.LastGC)
	MetricsMap["Lookups"] = float64(m.Lookups)
	MetricsMap["MCacheInuse"] = float64(m.MCacheInuse)
	MetricsMap["MCacheSys"] = float64(m.MCacheSys)
	MetricsMap["MSpanInuse"] = float64(m.MSpanInuse)
	MetricsMap["MSpanSys"] = float64(m.MSpanSys)
	MetricsMap["Mallocs"] = float64(m.Mallocs)
	MetricsMap["NextGC"] = float64(m.NextGC)
	MetricsMap["NumForcedGC"] = float64(m.NumForcedGC)
	MetricsMap["NumGC"] = float64(m.NumGC)
	MetricsMap["OtherSys"] = float64(m.OtherSys)
	MetricsMap["PauseTotalNs"] = float64(m.PauseTotalNs)
	MetricsMap["StackInuse"] = float64(m.StackInuse)
	MetricsMap["StackSys"] = float64(m.StackSys)
	MetricsMap["Sys"] = float64(m.Sys)
	MetricsMap["TotalAlloc"] = float64(m.TotalAlloc)
	MetricsMap["PollCount"] = 1
	MetricsMap["RandomValue"] = rand.Float64()
}

func UpdateMetricValue(metricType string, metricName string, value float64) {

	if metricType == "counter" {
		if metricName == "PollCount" {
			MetricsMap[metricName] += 1
		} else {
			MetricsMap[metricName] += value
		}
	} else {
		MetricsMap[metricName] = value
	}
}
func CheckMetricByName(metricName string) (float64, error) {
	value, exists := MetricsMap[metricName]
	if exists {
		return value, nil
	}
	return 0, fmt.Errorf("undefind metricName: %v", metricName)
}

func SortMetricByName() []string {
	var keys []string
	for key := range MetricsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
