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
}

func New() *Config {
	StorageConfig := storage.NewConfig(false)

	config := &Config{
		Storage:        StorageConfig,
		Address:        GetEnv("ADDRESS", "localhost:8080"),
		PollInterval:   GetEnvAsInt("POLL_INTERVAL", 2),
		ReportInterval: GetEnvAsInt("REPORT_INTERVAL", 10),
	}

	flag.StringVar(&config.Address, "a", GetEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.IntVar(&config.ReportInterval, "r", GetEnvAsInt("REPORT_INTERVAL", 10), "Frequency of sending metrics to the server")
	flag.IntVar(&config.PollInterval, "p", GetEnvAsInt("POLL_INTERVAL", 2), "Frequency of polling metrics from the runtime package")

	//flag.VisitAll(func(f *flag.Flag) {
	//	if f.Name == "a" || f.Name == "r" || f.Name == "p" {
	//		return
	//	}
	//
	//	fmt.Printf("Unknown agent flag: -%s\n", f.Name)
	//	flag.PrintDefaults()
	//	os.Exit(1)
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
