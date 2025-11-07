package models

import "fmt"

var (
	ErrEmptyMetricID       = fmt.Errorf("metric ID cannot be empty")
	ErrEmptyMetricType     = fmt.Errorf("metric type cannot be empty")
	ErrUnknownMetricType   = fmt.Errorf("unknown metric type")
	ErrValueRequired       = fmt.Errorf("value is required for gauge")
	ErrDeltaRequired       = fmt.Errorf("delta is required for counter")
	ErrInvalidCounterValue = fmt.Errorf("invalid counter value")
	ErrInvalidGaugeValue   = fmt.Errorf("invalid gauge value")
)
