package models

import (
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// Добавленный код для Metrics

// NewMetrics creates and validates a Metrics instance from raw values.
func NewMetrics(id, mType, valueStr string) (*Metrics, error) {
	if id == "" {
		return nil, ErrEmptyMetricID
	}
	if mType == "" {
		return nil, ErrEmptyMetricType
	}

	var metric Metrics
	metric.ID = id
	metric.MType = mType

	switch mType {
	case Counter:
		if valueStr == "" {
			return nil, ErrValueRequired
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return nil, ErrInvalidCounterValue
		}
		metric.Delta = &value

	case Gauge:
		if valueStr == "" {
			return nil, ErrValueRequired
		}
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, ErrInvalidGaugeValue
		}
		metric.Value = &value

	default:
		return nil, ErrUnknownMetricType
	}

	return &metric, nil
}

// Validate checks if the Metrics instance is valid.
func (m *Metrics) Validate() error {
	if m.ID == "" {
		return ErrEmptyMetricID
	}
	if m.MType == "" {
		return ErrEmptyMetricType
	}

	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return ErrDeltaRequired
		}
	case Gauge:
		if m.Value == nil {
			return ErrValueRequired
		}
	default:
		return ErrUnknownMetricType
	}

	return nil
}
