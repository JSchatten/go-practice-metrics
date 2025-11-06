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

func NewErrSaveToFile(err error) error {
	return fmt.Errorf("failed to save metrics to file: %w", err)
}

func NewErrLoadFromFile(err error) error {
	return fmt.Errorf("failed to load metrics from file: %w", err)
}

func NewErrMarshal(err error) error {
	return fmt.Errorf("failed to marshal metrics: %w", err)
}
