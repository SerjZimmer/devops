package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	Address        string
	PollInterval   int
	ReportInterval int
)

func FlagInit() {
	flag.StringVar(&Address, "a", getEnv("ADDRESS", "localhost:8080"), "Адрес эндпоинта HTTP-сервера")
	flag.IntVar(&ReportInterval, "r", getEnvAsInt("REPORT_INTERVAL", 10), "Частота отправки метрик на сервер")
	flag.IntVar(&PollInterval, "p", getEnvAsInt("POLL_INTERVAL", 2), "Частота опроса метрик из пакета runtime")
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "a" || f.Name == "r" || f.Name == "p" {
			return
		}
		fmt.Printf("Неизвестный флаг: -%s\n", f.Name)
		flag.PrintDefaults()
		os.Exit(1)
	})
	flag.Parse()
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
