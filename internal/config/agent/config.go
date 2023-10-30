package agent

import (
	"flag"
	"github.com/SerjZimmer/devops/internal/storage"
	"os"
	"strconv"
)

type Config struct {
	Address        string
	PollInterval   int
	ReportInterval int
	Storage        *storage.Config
	Key            string
}

func New() *Config {
	StorageConfig := storage.NewConfig(false)

	config := &Config{
		Storage:        StorageConfig,
		Address:        getEnv("ADDRESS", "localhost:8080"),
		PollInterval:   getEnvAsInt("POLL_INTERVAL", 2),
		ReportInterval: getEnvAsInt("REPORT_INTERVAL", 10),
	}

	flag.StringVar(&config.Address, "a", getEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.IntVar(&config.ReportInterval, "r", getEnvAsInt("REPORT_INTERVAL", 10), "Frequency of sending metrics to the server")
	flag.IntVar(&config.PollInterval, "p", getEnvAsInt("POLL_INTERVAL", 2), "Frequency of polling metrics from the runtime package")
	flag.StringVar(&config.Key, "k", getEnv("KEY", ""), "API Key for authentication")

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
