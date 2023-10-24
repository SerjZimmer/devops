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

		MaxConnections:  GetEnvAsInt("MAX_CONNECTIONS", 100),
		DatabaseDSN:     GetEnv("DATABASE_DSN", ""),
		RestoreFlag:     GetEnvAsBool("RESTORE", true),
		StoreInterval:   GetEnvAsInt("STORE_INTERVAL", 300),
		FileStoragePath: GetEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"),
	}
	flag.StringVar(&config.FileStoragePath, "f", GetEnv("FILE_STORAGE_PATH", "/tmp/metrics-db.json"), "Path to the file for storing metrics")
	//flag.StringVar(&config.Address, "a", GetEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")

	flag.IntVar(&config.MaxConnections, "c", GetEnvAsInt("MAX_CONNECTIONS", 100), "Maximum number of concurrent connections")
	flag.StringVar(&config.DatabaseDSN, "d", GetEnv("DATABASE_DSN", ""), "Database DSN")
	if needStoreInterval {
		flag.BoolVar(&config.RestoreFlag, "r", GetEnvAsBool("RESTORE", true), "Whether to restore previously saved metrics on server start")
	}
	flag.IntVar(&config.StoreInterval, "i", GetEnvAsInt("STORE_INTERVAL", 300), "Interval in seconds for storing server metrics on disk")

	//flag.VisitAll(func(f *flag.Flag) {
	//	if f.Name == "r" || f.Name == "i" || f.Name == "c" || f.Name == "d" || f.Name == "f" {
	//		return
	//	}
	//
	//	fmt.Printf("Unknown storage flag: -%s\n", f.Name)
	//	flag.PrintDefaults()
	//	os.Exit(1)
	//})
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
