package storage

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

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
	tempFile, err := ioutil.TempFile("", "testfile")
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
	content, err := ioutil.ReadFile(tempFile.Name())
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

func TestMetricsStorageInternal_ReadFromDisk(t *testing.T) {
	// Подготовка временного файла с данными
	tempFile, err := ioutil.TempFile("", "testfile.json")
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
	if err := ioutil.WriteFile(tempFile.Name(), testMetricsBytes, 0644); err != nil {
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
