package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/SerjZimmer/devops/internal/storage"
	"go.uber.org/zap"
)

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

// Handler представляет обработчик HTTP-запросов для взаимодействия с метриками.
type Handler struct {
	stor   metricsStorage
	logger *zap.Logger
}

// responseWriterWithStatus представляет ResponseWriter с поддержкой хранения HTTP-статуса.
type responseWriterWithStatus struct {
	http.ResponseWriter
	status int
}

// NewHandler создает новый экземпляр обработчика HTTP-запросов.
func NewHandler(stor metricsStorage) *Handler {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return &Handler{
		stor:   stor,
		logger: logger,
	}
}
