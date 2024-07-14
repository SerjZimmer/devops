package server

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"io/ioutil"
	"os"

	"github.com/SerjZimmer/devops/internal/storage"
)

// Config представляет структуру конфигурации для приложения.
type Config struct {
	Address   string
	LogLevel  string
	Storage   *storage.Config
	Key       string
	CryptoKey *rsa.PrivateKey
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

	cryptoKeyPath := getEnv("CRYPTO_KEY", "")
	flag.StringVar(&cryptoKeyPath, "crypto-key", getEnv("CRYPTO_KEY", ""), "Path to the private crypto key")
	flag.StringVar(&config.Address, "a", getEnv("ADDRESS", "localhost:8080"), "Address of the HTTP server endpoint")
	flag.StringVar(&config.LogLevel, "l", getEnv("LOG_LEVEL", "info"), "Logging level (e.g., 'info', 'debug')")
	flag.StringVar(&config.Key, "k", getEnv("KEY", ""), "API Key for authentication")
	flag.Parse()

	if cryptoKeyPath != "" {
		privKey, err := loadPrivateKey(cryptoKeyPath)
		if err != nil {
			panic(err)
		}
		config.CryptoKey = privKey
	}

	return config
}

// loadPrivateKey загружает приватный ключ из файла.
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey.(*rsa.PrivateKey), nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию, если переменная не установлена.
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}
