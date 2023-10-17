package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/SerjZimmer/devops/internal/config"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"
)

var metricKeys = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
	"PollCount",
	"RandomValue",
}

func createDB(DBConn string) {
	// Установка соединения с базой данных
	conn, err := pgx.Connect(context.Background(), DBConn)
	if err != nil {
		panic(err)
	}
	var tableExists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", "metrics").Scan(&tableExists)
	if err != nil {
		panic(err)
	}

	// Если таблица metrics не существует, создаем её
	if !tableExists {
		_, err = conn.Exec(context.Background(), `CREATE TABLE metrics (
            name text PRIMARY KEY,
            metric_data jsonb
        )`)
		if err != nil {
			panic(err)
		}
		fmt.Println("Таблица 'metrics' создана.")

		// Создайте записи с нулевыми значениями для каждого ключа в metricKeys
		for _, key := range metricKeys {
			delta := int64(0)
			value := 0.0
			metricData := Metrics{
				ID:    key,
				MType: "",
				Delta: &delta,
				Value: &value,
			}
			metricDataJSON, err := json.Marshal(metricData)
			if err != nil {
				fmt.Println(err)
				continue
			}

			_, err = conn.Exec(context.Background(), `
                INSERT INTO metrics (name, metric_data)
                VALUES ($1, $2)
                ON CONFLICT (name) DO UPDATE
                SET metric_data = $2
            `, key, metricDataJSON)

			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println("Таблица 'metrics' уже существует.")
	}
}

type MetricsStorage struct {
	Mu         sync.RWMutex
	MetricsMap map[string]float64
	c          *config.Config
	DB         *sql.DB
}

func TestMetricStorage() *MetricsStorage {
	m := &MetricsStorage{
		MetricsMap: make(map[string]float64),
	}

	return m
}

func (s *MetricsStorage) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.DB.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func NewMetricsStorage(c *config.Config) *MetricsStorage {
	db, err := sql.Open("pgx", c.DatabaseDSN)
	if err != nil {
		panic(err)
	}

	createDB(c.DatabaseDSN)

	m := &MetricsStorage{
		MetricsMap: make(map[string]float64),
		c:          c,
		DB:         db,
	}

	if c.RestoreFlag {
		_ = m.ReadFromDisk()
	}
	go func() {
		if c.StoreInterval > 0 {
			t := time.NewTicker(time.Duration(c.StoreInterval) * time.Second)
			for i := range t.C {
				err = m.writeToDisk()
				if err != nil {
					fmt.Println(i)
					fmt.Println(err)
				}
			}
		}

	}()
	return m
}
func (s *MetricsStorage) ReadFromDisk() error {
	bytes, err := os.ReadFile(s.c.FileStoragePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &s.MetricsMap)
}

func (s *MetricsStorage) Shutdown() {
	_ = s.writeToDisk()
}

func (s *MetricsStorage) writeToDisk() error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	copyMetricsMap := make(map[string]float64)
	for key, value := range s.MetricsMap {
		copyMetricsMap[key] = value
	}

	bytes, err := json.Marshal(copyMetricsMap)
	if err != nil {
		return err
	}

	return os.WriteFile(s.c.FileStoragePath, bytes, 0644)
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
	if s.c.StoreInterval == 0 {
		_ = s.writeToDisk()
	}
}
func keyExists(key string) bool {
	for _, k := range metricKeys {
		if k == key {
			return true
		}
	}
	return false
}
func (s *MetricsStorage) UpdateMetricValue(m Metrics) {
	s.Mu.Lock()

	if m.MType == "counter" {
		if m.Delta == nil {
			v := int64(1)
			m.Delta = &v
		}
		s.MetricsMap[m.ID] += float64(*m.Delta)

		d := int64(s.MetricsMap[m.ID])
		metricData := Metrics{
			ID:    m.ID,
			MType: m.MType,
			Delta: &d,
			Value: m.Value,
		}

		metricDataJSON, err := json.Marshal(metricData)
		if err != nil {
			fmt.Println(err)
			s.Mu.Unlock()
		}
		if keyExists(m.ID) == false {
			_, err := s.DB.ExecContext(context.Background(), "INSERT INTO metrics (name, metric_data) VALUES ($1, $2)", m.ID, metricDataJSON)
			if err != nil {
				fmt.Println(err)
			}
		}
		_, err = s.DB.ExecContext(context.Background(), "UPDATE metrics SET metric_data = $1 WHERE name = $2", metricDataJSON, m.ID)
		if err != nil {
			fmt.Println(err)
		}

	} else {
		d := int64(0)
		s.MetricsMap[m.ID] = *m.Value

		metricData := Metrics{
			ID:    m.ID,
			MType: m.MType,
			Delta: &d,
			Value: m.Value,
		}

		metricDataJSON, err := json.Marshal(metricData)
		if err != nil {
			fmt.Println(err)
			s.Mu.Unlock()
		}
		if keyExists(m.ID) == false {
			_, err := s.DB.ExecContext(context.Background(), "INSERT INTO metrics (name, metric_data) VALUES ($1, $2)", m.ID, metricDataJSON)
			if err != nil {
				fmt.Println(err)
			}
		}

		_, err = s.DB.ExecContext(context.Background(), "UPDATE metrics SET metric_data = $1 WHERE name = $2", metricDataJSON, m.ID)
		if err != nil {
			fmt.Println(err)
		}
	}

	s.Mu.Unlock()
}

func (s *MetricsStorage) GetMetricByName(m Metrics) (float64, error) {
	s.Mu.RLock()
	value, exists := s.MetricsMap[m.ID]
	s.Mu.RUnlock()
	if exists {
		return value, nil
	}
	return 0, fmt.Errorf("undefind metricName: %v", m.ID)
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
