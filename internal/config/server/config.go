package server

import (
	"flag"
	"os"

	"github.com/SerjZimmer/devops/internal/storage"
)

// Config представляет структуру конфигурации для приложения.
type Config struct {
	Address  string
	LogLevel string
	Storage  *storage.Config
	Key      string
}

// New создает новый экземпляр конфигурации с значениями по умолчанию или из переменных окружения и флагов командной строки.
func New() *Config {
	StorageConfig := storage.NewConfig(true)
	config := &Config{
		Address:  getEnv("ADDRESS", "localhost:8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Storage:  StorageConfig,
		Key:      getEnv("KEY", ""),
	}

	flag.StringVar(&config.Address, "a", getEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.StringVar(&config.LogLevel, "l", getEnv("LOG_LEVEL", "info"), "Logging level (e.g., 'info', 'debug')")
	flag.StringVar(&config.Key, "k", getEnv("KEY", ""), "API Key for authentication")
	flag.Parse()
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
