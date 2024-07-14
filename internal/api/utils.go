package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"strconv"

	"github.com/SerjZimmer/devops/internal/storage"
)

// Handler представляет обработчик HTTP-запросов.
type Handler struct {
	store     *storage.MetricsStorage
	CryptoKey *rsa.PrivateKey
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(storage *storage.MetricsStorage, cryptoKey *rsa.PrivateKey) *Handler {
	return &Handler{
		store:     storage,
		CryptoKey: cryptoKey,
	}
}

// decryptMessage дешифрует сообщение с использованием приватного ключа.
func decryptMessage(privKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, ciphertext, nil)
}

// metricsListTemplate - HTML-шаблон для представления списка метрик.
const metricsListTemplate = `
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Все метрики</h1>
    <ul>
        {{range .Metrics}}
        <li>{{.}}</li>
        {{end}}
    </ul>
</body>
</html>
`

var tmpl = template.Must(template.New("metricsList").Parse(metricsListTemplate))

// isValidMetrics проверяет корректность переданных данных метрик.
func isValidMetrics(m storage.Metrics) bool {
	if m.ID == "" {
		return false
	}

	if m.MType != "gauge" && m.MType != "counter" {
		return false
	}

	if m.MType == "gauge" && m.Value == nil {
		return false
	}

	if m.MType == "counter" && m.Delta == nil && m.ID != "PollCount" {
		return false
	}

	return true
}

// parseNumeric преобразует строковое значение в числовой формат.
func parseNumeric(mValue string) (float64, error) {
	floatVal, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		return 0, err
	}
	return floatVal, nil
}

// calculateHash вычисляет хеш SHA256 для переданных данных.
func calculateHash(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash
}

// responseWriterWithStatus представляет ResponseWriter с поддержкой хранения HTTP-статуса.
type responseWriterWithStatus struct {
	http.ResponseWriter
	status int
}
