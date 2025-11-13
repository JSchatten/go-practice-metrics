package service

import (
	"errors"
	"fmt"
)

// Общие ошибки
var (
	ErrUnknownMetricType        = errors.New("unknown metric type")
	ErrValueRequired            = errors.New("value is required for gauge metric")
	ErrDeltaRequired            = errors.New("delta is required for counter metric")
	ErrMetricNotFound           = errors.New("metric not found")
	ErrMetricSaveFailedDatabase = errors.New("failed to save metric to database")
	ErrMetricSaveFailedFile     = errors.New("failed to save metric to file")
	ErrMetricSaveMemory         = errors.New("failed to save metric to memory")
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
