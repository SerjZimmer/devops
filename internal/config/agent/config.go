package agent

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"io/ioutil"
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
	CryptoKey      *rsa.PublicKey
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

	cryptoKeyPath := getEnv("CRYPTO_KEY", "")
	flag.StringVar(&cryptoKeyPath, "crypto-key", getEnv("CRYPTO_KEY", ""), "Path to the public crypto key")
	flag.StringVar(&config.Address, "a", getEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.IntVar(&config.ReportInterval, "r", getEnvAsInt("REPORT_INTERVAL", 10), "Frequency of sending metrics to the server")
	flag.IntVar(&config.PollInterval, "p", getEnvAsInt("POLL_INTERVAL", 2), "Frequency of polling metrics from the runtime package")
	flag.StringVar(&config.Key, "k", getEnv("KEY", ""), "API Key for authentication")
	flag.IntVar(&config.RateLimit, "l", getEnvAsInt("RATE_LIMIT", 1), "Rate limit value")

	flag.Parse()

	if cryptoKeyPath != "" {
		pubKey, err := loadPublicKey(cryptoKeyPath)
		if err != nil {
			panic(err)
		}
		config.CryptoKey = pubKey
	}

	return config
}

// loadPublicKey загружает публичный ключ из файла.
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubKey.(*rsa.PublicKey), nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию, если переменная не установлена.
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt возвращает значение переменной окружения в виде целого числа или значение по умолчанию, если переменная не установлена.
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
