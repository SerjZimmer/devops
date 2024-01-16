package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"sync"
	"testing"
)

func TestWriteMetrics(t *testing.T) {
	// Создаем объект MetricsStorageInternal
	metricsStorage := &MetricsStorageInternal{
		MetricsMap: make(map[string]float64),
		Mu:         sync.RWMutex{},
		c: &Config{
			StoreInterval: 0, // установка StoreInterval в 0 для вызова writeToDisk внутри функции
		},
		DB: &sql.DB{}, // Предполагая, что у вас есть тип sql.DB
	}

	// Инициализируем структуру MemStats для передачи в WriteMetrics
	memStats := runtime.MemStats{
		Alloc:       100,
		BuckHashSys: 200,
	}

	// Вызываем функцию WriteMetrics
	metricsStorage.WriteMetrics(memStats)

	// Проверяем, что метрики были корректно записаны в MetricsMap
	assert.Equal(t, float64(100), metricsStorage.MetricsMap["Alloc"])
	assert.Equal(t, float64(200), metricsStorage.MetricsMap["BuckHashSys"])

	// Проверяем, что PollCount был установлен в 1
	assert.Equal(t, float64(1), metricsStorage.MetricsMap["PollCount"])

	// Проверяем, что RandomValue был установлен (в пределах разумного)
	assert.NotNil(t, metricsStorage.MetricsMap["RandomValue"])
}

func TestCollectNewMetrics(t *testing.T) {
	// Создание объекта MetricsStorageInternal
	metricsStorage := &MetricsStorageInternal{
		MetricsMap: make(map[string]float64),
		Mu:         sync.RWMutex{},
		c:          &Config{}, // Предполагая, что у вас есть тип Config
		DB:         &sql.DB{}, // Предполагая, что у вас есть тип sql.DB
	}

	// Тест для случая успешного сбора метрик
	t.Run("Collect New Metrics Successfully", func(t *testing.T) {
		// Вызываем функцию для сбора метрик
		collectNewMetrics(metricsStorage)

		// Проверяем, что метрики были успешно добавлены в MetricsMap
		assert.GreaterOrEqual(t, metricsStorage.MetricsMap["TotalMemory"], float64(0))
		assert.GreaterOrEqual(t, metricsStorage.MetricsMap["FreeMemory"], float64(0))

		key := fmt.Sprintf("CPUUtilization%d", 0)
		assert.GreaterOrEqual(t, metricsStorage.MetricsMap[key], float64(0))

	})

}
func TestNewConfig(t *testing.T) {
	// Тест для создания нового Config с значениями по умолчанию
	config := NewConfig(false)

	assert.Equal(t, 100, config.MaxConnections)
	assert.Equal(t, "", config.DatabaseDSN)
	assert.Equal(t, true, config.RestoreFlag)
	assert.Equal(t, 300, config.StoreInterval)
	assert.Equal(t, "/tmp/metrics-db.json", config.FileStoragePath)
}

func TestGetEnvAsInt(t *testing.T) {
	// Тест для получения значения переменной окружения как int
	os.Setenv("TEST_INT_ENV", "42")
	defer os.Unsetenv("TEST_INT_ENV")

	result := getEnvAsInt("TEST_INT_ENV", 0)

	assert.Equal(t, 42, result)
}

func TestGetEnvAsBool(t *testing.T) {
	// Тест для получения значения переменной окружения как bool
	os.Setenv("TEST_BOOL_ENV", "true")
	defer os.Unsetenv("TEST_BOOL_ENV")

	result := getEnvAsBool("TEST_BOOL_ENV", false)

	assert.Equal(t, true, result)
}

func TestUpdateMetricValue(t *testing.T) {
	// Инициализация MetricsStorageInternal
	storage := &MetricsStorageInternal{
		MetricsMap: make(map[string]float64),
	}

	// Тест для counter с Delta равным nil
	t.Run("UpdateMetricValue Counter with Nil Delta", func(t *testing.T) {
		metrics := Metrics{
			ID:    "metric1",
			MType: "counter",
		}

		err := storage.UpdateMetricValue(metrics)

		assert.NoError(t, err)
		assert.Equal(t, float64(1), storage.MetricsMap["metric1"])
	})

	// Тест для counter с указанным Delta
	t.Run("UpdateMetricValue Counter with Non-nil Delta", func(t *testing.T) {
		// Предварительная установка значения метрики
		storage.MetricsMap["metric2"] = 5.0

		metrics := Metrics{
			ID:    "metric2",
			MType: "counter",
			Delta: int64Ptr(3),
		}

		err := storage.UpdateMetricValue(metrics)

		assert.NoError(t, err)
		assert.Equal(t, float64(8), storage.MetricsMap["metric2"])
	})

	// Тестирование JSON маршалинга
	t.Run("UpdateMetricValue JSON Marshaling", func(t *testing.T) {
		metrics := Metrics{
			ID:    "metric3",
			MType: "counter",
			Delta: int64Ptr(2),
			Value: float64Ptr(10.5),
		}

		err := storage.UpdateMetricValue(metrics)

		assert.NoError(t, err)

		// Проверка, что Delta преобразуется в JSON корректно
		expectedJSON := `{"id":"metric3","type":"counter","delta":2,"value":10.5}`
		actualJSON, err := json.Marshal(metrics)

		assert.NoError(t, err)
		assert.Equal(t, expectedJSON, string(actualJSON))
	})
	t.Run("UpdateMetricValue Non-Counter Type", func(t *testing.T) {
		metrics := Metrics{
			ID:    "metric5",
			MType: "gauge",
			Value: float64Ptr(7.5),
		}

		err := storage.UpdateMetricValue(metrics)

		assert.NoError(t, err)
		assert.Equal(t, float64(7.5), storage.MetricsMap["metric5"])
	})

	// Тест для случая ошибки при маршалинге JSON (вторая часть функции)
	t.Run("UpdateMetricValue JSON Marshaling Error (Part 2)", func(t *testing.T) {
		// Намеренно создаем ошибку маршалинга
		metrics := Metrics{
			ID:    "metric6",
			MType: "gauge",
			Value: float64Ptr(12.3),
		}

		err := storage.UpdateMetricValue(metrics)

		assert.NoError(t, err)

		// Проверка, что Delta преобразуется в JSON корректно
		expectedJSON := `{"id":"metric6","type":"gauge","value":12.3}`
		actualJSON, err := json.Marshal(metrics)

		assert.NoError(t, err)
		assert.Equal(t, expectedJSON, string(actualJSON))
	})
}
func int64Ptr(v int64) *int64 {
	return &v
}

// Вспомогательная функция для создания указателя на float64
func float64Ptr(v float64) *float64 {
	return &v
}
func TestKeyExists(t *testing.T) {
	// Тест случая, когда ключ существует
	existingKey := "HeapAlloc"
	assert.True(t, keyExists(existingKey), "Expected KeyExists to return true for existing key, but got false")

	// Тест случая, когда ключ отсутствует
	nonExistingKey := "NonExistingKey"
	assert.False(t, keyExists(nonExistingKey), "Expected KeyExists to return false for non-existing key, but got true")

	// Тест случая, когда ключ добавляется в массив
	newKey := "NewKey"
	assert.False(t, keyExists(newKey), "Expected KeyExists to return false for new key before it's added, but got true")
	assert.True(t, keyExists(newKey), "Expected KeyExists to return true for new key after it's added, but got false")
}

// FileSystem интерфейс для работы с файловой системой.
type FileSystem interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// DefaultFileSystem реализует FileSystem интерфейс, используя стандартные функции пакета os.
type DefaultFileSystem struct{}

// WriteFile реализует запись файла с использованием стандартной функции пакета os.
func (fs DefaultFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// Тест для функции writeToDisk
func TestWriteToDisk(t *testing.T) {
	// Создаем временный файл
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Создаем экземпляр хранилища метрик для использования в тестах.
	storage := &MetricsStorageInternal{
		MetricsMap: make(map[string]float64),
		c: &Config{
			FileStoragePath: tempFile.Name(),
		},
	}

	// Добавляем тестовые данные
	storage.MetricsMap["metric1"] = 123.45
	storage.MetricsMap["metric2"] = 678.90

	// Записываем в файл
	err = storage.writeToDisk()
	if err != nil {
		t.Fatalf("writeToDisk failed: %v", err)
	}

	// Читаем данные из файла
	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Could not read from temp file: %v", err)
	}

	// Распаковываем данные и сравниваем
	var readMetrics map[string]float64
	err = json.Unmarshal(content, &readMetrics)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Проверяем, что прочитанные метрики совпадают с ожидаемыми
	if len(storage.MetricsMap) != len(readMetrics) {
		t.Errorf("Number of metrics does not match. Expected: %v, Got: %v", len(storage.MetricsMap), len(readMetrics))
	}

	for key, expectedValue := range storage.MetricsMap {
		if actualValue, ok := readMetrics[key]; ok {
			if expectedValue != actualValue {
				t.Errorf("Metric value mismatch for key %s. Expected: %v, Got: %v", key, expectedValue, actualValue)
			}
		} else {
			t.Errorf("Key %s not found in read metrics", key)
		}
	}
}

// t
func TestMetricsStorageInternal_ReadFromDisk(t *testing.T) {
	// Подготовка временного файла с данными
	tempFile, err := os.CreateTemp("", "testfile.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Подготовка тестовых данных
	testMetrics := map[string]float64{"metric1": 123, "metric2": 0}
	testMetricsBytes, err := json.Marshal(testMetrics)
	if err != nil {
		t.Fatal(err)
	}

	// Запись данных во временный файл
	if err := os.WriteFile(tempFile.Name(), testMetricsBytes, 0644); err != nil {
		t.Fatal(err)
	}

	// Подготовка объекта MetricsStorageInternal для теста
	storage := MetricsStorageInternal{
		c: &Config{FileStoragePath: tempFile.Name()},
		// Добавьте необходимые инициализации для теста
	}

	// Вызов тестируемой функции
	err = storage.ReadFromDisk()

	// Проверка результата
	assert.NoError(t, err)
	assert.Equal(t, testMetrics, storage.MetricsMap)
}

func TestMetricsStorageInternal_GetMetricByName(t *testing.T) {
	// Подготовка тестовых данных
	testMetricsMap := map[string]float64{
		"metric1": 123,
		"metric2": 456,
	}
	config := &Config{} // Добавьте необходимые данные в конфигурацию

	// Инициализация объекта MetricsStorageInternal для теста
	storage := MetricsStorageInternal{
		Mu:         sync.RWMutex{},
		MetricsMap: testMetricsMap,
		c:          config,
		// Добавьте необходимые инициализации для теста
	}

	// Вызов тестируемой функции
	result, err := storage.GetMetricByName(Metrics{ID: "metric1"})

	// Проверка результата
	assert.NoError(t, err)
	assert.Equal(t, 123.0, result)

	// Проверка сценария с несуществующей метрикой
	result, err = storage.GetMetricByName(Metrics{ID: "nonexistent_metric"})

	// Проверка результата
	assert.Error(t, err)
	assert.EqualError(t, err, "undefind metricName: nonexistent_metric")
	assert.Equal(t, 0.0, result)
}

func TestMetricsStorageInternal_SortMetricByName(t *testing.T) {
	// Подготовка тестовых данных
	testMetricsMap := map[string]float64{
		"metric1": 123,
		"metric3": 456,
		"metric2": 789,
	}
	config := &Config{} // Добавьте необходимые данные в конфигурацию

	// Инициализация объекта MetricsStorageInternal для теста
	storage := MetricsStorageInternal{
		Mu:         sync.RWMutex{},
		MetricsMap: testMetricsMap,
		c:          config,
		// Добавьте необходимые инициализации для теста
	}

	// Вызов тестируемой функции
	result := storage.SortMetricByName()

	// Ожидаемый порядок после сортировки
	expectedOrder := []string{"metric1", "metric2", "metric3"}

	// Проверка результата
	assert.Equal(t, expectedOrder, result)
}

func TestMetricsStorageInternal_GetAllMetrics(t *testing.T) {
	// Подготовка тестовых данных
	testMetricsMap := map[string]float64{
		"metric1": 123,
		"metric3": 456,
		"metric2": 789,
	}
	config := &Config{} // Добавьте необходимые данные в конфигурацию

	// Инициализация объекта MetricsStorageInternal для теста
	storage := MetricsStorageInternal{
		Mu:         sync.RWMutex{},
		MetricsMap: testMetricsMap,
		c:          config,
		// Добавьте необходимые инициализации для теста
	}

	// Вызов тестируемой функции
	result := storage.GetAllMetrics()

	// Ожидаемый результат
	expectedResult := "metric1/123\nmetric2/789\nmetric3/456\n"

	// Проверка результата
	assert.Equal(t, expectedResult, result)
}
