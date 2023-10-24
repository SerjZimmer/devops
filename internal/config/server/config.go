package server

import (
	"flag"
	"github.com/SerjZimmer/devops/internal/storage"
	"os"
)

type Config struct {
	Address  string
	LogLevel string
	Storage  *storage.Config
}

func New() *Config {
	StorageConfig := storage.NewConfig(true)
	config := &Config{
		Address:  getEnv("ADDRESS", "localhost:8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Storage:  StorageConfig,
	}
	flag.StringVar(&config.Address, "a", getEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.StringVar(&config.LogLevel, "l", getEnv("LOG_LEVEL", "info"), "Logging level (e.g., 'info', 'debug')")
	flag.Parse()
	return config
}
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}
