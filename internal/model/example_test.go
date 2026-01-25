package models_test

import (
	"fmt"
	"log"

	models "github.com/JSchatten/go-practice-metrics/internal/model"
)

// ExampleNewMetrics_Counter демонстрирует создание счётчика (counter).
// Показывает успешное создание и базовую валидацию.
func ExampleNewMetrics_counter() {
	metric, err := models.NewMetrics("poll_count", "counter", "42")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s\n", metric.ID)
	fmt.Printf("Type: %s\n", metric.MType)
	fmt.Printf("Delta: %d\n", *metric.Delta)
	// Output:
	// ID: poll_count
	// Type: counter
	// Delta: 42
}

// ExampleNewMetrics_Gauge демонстрирует создание измерителя (gauge).
func ExampleNewMetrics_gauge() {
	metric, err := models.NewMetrics("cpu_load", "gauge", "0.75")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s\n", metric.ID)
	fmt.Printf("Type: %s\n", metric.MType)
	fmt.Printf("Value: %.2f\n", *metric.Value)
	// Output:
	// ID: cpu_load
	// Type: gauge
	// Value: 0.75
}

// ExampleNewMetrics_invalidType демонстрирует обработку ошибки при указании неизвестного типа.
func ExampleNewMetrics_invalidType() {
	_, err := models.NewMetrics("unknown", "timer", "100")
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: unknown metric type
}

// ExampleNewMetrics_emptyID демонстрирует ошибку при пустом имени метрики.
func ExampleNewMetrics_emptyID() {
	_, err := models.NewMetrics("", "gauge", "0.5")
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: metric ID cannot be empty
}

// ExampleMetrics_Validate_validCounter демонстрирует валидацию корректной метрики типа counter.
func ExampleMetrics_Validate_validCounter() {
	value := int64(100)
	metric := &models.Metrics{
		ID:    "requests",
		MType: models.Counter,
		Delta: &value,
	}

	err := metric.Validate()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Validation passed")
	// Output:
	// Validation passed
}

// ExampleMetrics_Validate_invalidGauge демонстрирует ошибку валидации при отсутствии значения у gauge.
func ExampleMetrics_Validate_invalidGauge() {
	metric := &models.Metrics{
		ID:    "temperature",
		MType: models.Gauge,
		// Value не задано - это ошибка
	}

	err := metric.Validate()
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: value is required for gauge
}

// ExampleMetrics_String демонстрирует строковое представление метрики в формате JSON.
func ExampleMetrics_String() {
	value := float64(0.95)
	metric := &models.Metrics{
		ID:    "efficiency",
		MType: models.Gauge,
		Value: &value,
		Hash:  "abc123",
	}

	fmt.Println(metric)
	// Output:
	// {"id":"efficiency","type":"gauge","value":0.95,"hash":"abc123"}
}
