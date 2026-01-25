package service

import (
	"errors"
	"fmt"
)

// Общие ошибки, возвращаемые методами хранилища.
// Используются для согласованной обработки сбоев в различных слоях приложения.
var (
	// ErrUnknownMetricType возвращается, когда указан неизвестный тип метрики
	// (не "gauge" и не "counter").
	ErrUnknownMetricType = errors.New("unknown metric type")

	// ErrValueRequired возвращается, если для метрики типа gauge не указано значение.
	ErrValueRequired = errors.New("value is required for gauge metric")

	// ErrDeltaRequired возвращается, если для метрики типа counter не указано delta.
	ErrDeltaRequired = errors.New("delta is required for counter metric")

	// ErrMetricNotFound возвращается, когда метрика с указанным именем не найдена.
	ErrMetricNotFound = errors.New("metric not found")

	// ErrMetricSaveFailedDatabase возвращается при ошибке сохранения метрики в БД.
	ErrMetricSaveFailedDatabase = errors.New("failed to save metric to database")

	// ErrMetricSaveFailedFile возвращается при ошибке записи метрик в файл.
	ErrMetricSaveFailedFile = errors.New("failed to save metric to file")

	// ErrMetricSaveMemory возвращается при ошибке обновления метрики в памяти.
	ErrMetricSaveMemory = errors.New("failed to save metric to memory")
)

// Форматированные ошибки
func NewErrUnknownMetricType(mtype string) error {
	return fmt.Errorf("%w: %s", ErrUnknownMetricType, mtype)
}

func NewErrSaveToFile(filename string, err error) error {
	return fmt.Errorf("failed to save metrics to file %s: %w", filename, err)
}

func NewErrLoadFromFile(filename string, err error) error {
	return fmt.Errorf("failed to load metrics from file %s: %w", filename, err)
}

func NewErrMarshal(err error) error {
	return fmt.Errorf("failed to marshal metrics: %w", err)
}
