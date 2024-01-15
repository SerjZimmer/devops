package storage

import (
	"flag"
	"os"
	"strconv"
)

// Config представляет собой структуру конфигурации для хранилища метрик.
type Config struct {
	RestoreFlag     bool
	MaxConnections  int
	DatabaseDSN     string
	StoreInterval   int
	FileStoragePath string
}

// NewConfig создает новый экземпляр конфигурации хранилища метрик.
func NewConfig(needStoreInterval bool) *Config {

	config := &Config{

		MaxConnections:  getEnvAsInt("MAX_CONNECTIONS", 100),
		DatabaseDSN:     getEnv("DATABASE_DSN", ""),
		RestoreFlag:     getEnvAsBool("RESTORE", true),
		StoreInterval:   getEnvAsInt("STORE_INTERVAL", 300),
		FileStoragePath: getEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"),
	}
	flag.StringVar(&config.FileStoragePath, "f", getEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"), "Path to the file for storing metrics")
	flag.IntVar(&config.MaxConnections, "c", getEnvAsInt("MAX_CONNECTIONS", 100), "Maximum number of concurrent connections")
	flag.StringVar(&config.DatabaseDSN, "d", getEnv("DATABASE_DSN", ""), "Database DSN")
	if needStoreInterval {
		flag.BoolVar(&config.RestoreFlag, "r", getEnvAsBool("RESTORE", true), "Whether to restore previously saved metrics on server start")
	}
	flag.IntVar(&config.StoreInterval, "i", getEnvAsInt("STORE_INTERVAL", 300), "Interval in seconds for storing server metrics on disk")
	return config
}

// getEnv возвращает значение переменной окружения или значение по умолчанию, если переменная не установлена.
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt возвращает значение переменной окружения в виде целого числа или значение по умолчанию, если переменная не установлена или не является числом.
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

// getEnvAsBool возвращает значение переменной окружения в виде булевого значения или значение по умолчанию, если переменная не установлена или не является булевым значением.
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr != "" {
		value, err := strconv.ParseBool(valueStr)
		if err == nil {
			return value
		}
	}
	return defaultValue
}
