package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Address         string
	PollInterval    int
	ReportInterval  int
	StoreInterval   int
	FileStoragePath string
	RestoreFlag     bool
	LogLevel        string
	MaxConnections  int
}

func NewConfig() *Config {
	config := &Config{
		Address:         getEnv("ADDRESS", "localhost:8080"),
		PollInterval:    getEnvAsInt("POLL_INTERVAL", 2),
		ReportInterval:  getEnvAsInt("REPORT_INTERVAL", 10),
		StoreInterval:   getEnvAsInt("STORE_INTERVAL", 300),
		FileStoragePath: getEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"),
		RestoreFlag:     getEnvAsBool("RESTORE", true),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		MaxConnections:  getEnvAsInt("MAX_CONNECTIONS", 100),
	}

	flag.StringVar(&config.Address, "a", getEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.IntVar(&config.ReportInterval, "r", getEnvAsInt("REPORT_INTERVAL", 10), "Frequency of sending metrics to the server")
	flag.IntVar(&config.PollInterval, "p", getEnvAsInt("POLL_INTERVAL", 2), "Frequency of polling metrics from the runtime package")
	flag.IntVar(&config.StoreInterval, "i", getEnvAsInt("STORE_INTERVAL", 300), "Interval in seconds for storing server metrics on disk")
	flag.StringVar(&config.FileStoragePath, "f", getEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"), "Path to the file for storing metrics")
	//flag.BoolVar(&config.RestoreFlag, "b", getEnvAsBool("RESTORE", true), "Whether to restore previously saved metrics on server start")
	flag.StringVar(&config.LogLevel, "l", getEnv("LOG_LEVEL", "info"), "Logging level (e.g., 'info', 'debug')")
	flag.IntVar(&config.MaxConnections, "c", getEnvAsInt("MAX_CONNECTIONS", 100), "Maximum number of concurrent connections")

	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "a" || f.Name == "p" || f.Name == "i" || f.Name == "f" || f.Name == "r" || f.Name == "l" || f.Name == "c" {
			return
		}

		fmt.Printf("Unknown flag: -%s\n", f.Name)
		flag.PrintDefaults()
		os.Exit(1)
	})
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
