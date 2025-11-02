package config

import "fmt"

var (
	ErrInvalidPollInterval   = fmt.Errorf("error: pollInterval must be positive")
	ErrInvalidReportInterval = fmt.Errorf("error: reportInterval must be positive")
)
