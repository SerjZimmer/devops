package agent

import (
	"flag"
	"os"
	"strconv"

	"github.com/SerjZimmer/devops/internal/storage"
)

// Config представляет структуру конфигурации для приложения.
type Config struct {
	Address        string
	PollInterval   int
	ReportInterval int
	Storage        *storage.Config
	Key            string
	RateLimit      int
}

// New создает новый экземпляр конфигурации с значениями по умолчанию или из переменных окружения и флагов командной строки.
func New() *Config {
	StorageConfig := storage.NewConfig(false)

	config := &Config{
		Storage:        StorageConfig,
		Address:        getEnv("ADDRESS", "localhost:8080"),
		PollInterval:   getEnvAsInt("POLL_INTERVAL", 2),
		ReportInterval: getEnvAsInt("REPORT_INTERVAL", 10),
		Key:            getEnv("KEY", ""),
		RateLimit:      getEnvAsInt("RATE_LIMIT", 1),
	}

	flag.StringVar(&config.Address, "a", getEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.IntVar(&config.ReportInterval, "r", getEnvAsInt("REPORT_INTERVAL", 10), "Frequency of sending metrics to the server")
	flag.IntVar(&config.PollInterval, "p", getEnvAsInt("POLL_INTERVAL", 2), "Frequency of polling metrics from the runtime package")
	flag.StringVar(&config.Key, "k", getEnv("KEY", ""), "API Key for authentication")
	flag.IntVar(&config.RateLimit, "l", getEnvAsInt("RATE_LIMIT", 1), "Rate limit value")

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

// getEnvAsInt возвращает значение переменной окружения в виде целого числа или значение по умолчанию, если переменная не установлена
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
