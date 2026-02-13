// Package models содержит определение ошибок, используемых в рамках пакета модели метрик.
// Все ошибки, связанные с валидацией и созданием метрик, объявлены здесь.
package model

import "errors"

var (
	ErrEmptyMetricID       = errors.New("metric ID cannot be empty")
	ErrEmptyMetricType     = errors.New("metric type cannot be empty")
	ErrUnknownMetricType   = errors.New("unknown metric type")
	ErrValueRequired       = errors.New("value is required for gauge")
	ErrDeltaRequired       = errors.New("delta is required for counter")
	ErrInvalidCounterValue = errors.New("invalid counter value")
	ErrInvalidGaugeValue   = errors.New("invalid gauge value")
	ErrNilMetric           = errors.New("metric is nil")
)
