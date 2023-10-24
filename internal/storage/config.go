package storage

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	RestoreFlag     bool
	MaxConnections  int
	DatabaseDSN     string
	StoreInterval   int
	FileStoragePath string
}

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
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}

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
