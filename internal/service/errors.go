package service

import "fmt"

// Общие ошибки
var (
	ErrUnknownMetricType = fmt.Errorf("unknown metric type")
	ErrValueRequired     = fmt.Errorf("value is required for gauge metric")
	ErrDeltaRequired     = fmt.Errorf("delta is required for counter metric")
	ErrMetricNotFound    = fmt.Errorf("metric not found")
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
