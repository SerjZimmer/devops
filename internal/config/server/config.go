package server

import (
	"flag"
	"github.com/SerjZimmer/devops/internal/storage"
	"os"
	"strconv"
)

type Config struct {
	Address         string
	StoreInterval   int
	FileStoragePath string
	RestoreFlag     bool
	LogLevel        string
	Storage         *storage.Config
}

func New() *Config {
	StorageConfig := storage.NewConfig(true)
	config := &Config{
		Address: GetEnv("ADDRESS", "localhost:8080"),
		//StoreInterval: GetEnvAsInt("STORE_INTERVAL", 300),
		//FileStoragePath: GetEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"),
		//RestoreFlag: GetEnvAsBool("RESTORE", true),
		LogLevel: GetEnv("LOG_LEVEL", "info"),
		Storage:  StorageConfig,
	}

	flag.StringVar(&config.Address, "a", GetEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	///flag.IntVar(&config.StoreInterval, "i", GetEnvAsInt("STORE_INTERVAL", 300), "Interval in seconds for storing server metrics on disk")
	//flag.StringVar(&config.FileStoragePath, "f", GetEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"), "Path to the file for storing metrics")
	//flag.BoolVar(&config.RestoreFlag, "r", GetEnvAsBool("RESTORE", true), "Whether to restore previously saved metrics on server start")
	flag.StringVar(&config.LogLevel, "l", GetEnv("LOG_LEVEL", "info"), "Logging level (e.g., 'info', 'debug')")

	//flag.VisitAll(func(f *flag.Flag) {
	//	//if f.Name == "a" || f.Name == "p" || f.Name == "r" || f.Name == "i" || f.Name == "f" || f.Name == "b" || f.Name == "l" || f.Name == "c" || f.Name == "d" {
	//	//	return
	//	//}
	//	//
	//	//fmt.Printf("Unknown server flag: -%s\n", f.Name)
	//	//flag.PrintDefaults()
	//	//os.Exit(1)
	//})
	flag.Parse()
	return config
}
func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err == nil {
			return value
		}
	}
	return defaultValue
}
func GetEnvAsBool(key string, defaultValue bool) bool {
	valueStr := GetEnv(key, "")
	if valueStr != "" {
		value, err := strconv.ParseBool(valueStr)
		if err == nil {
			return value
		}
	}
	return defaultValue
}
