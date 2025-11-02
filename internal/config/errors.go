package config

import "fmt"

var (
	ErrInvalidPollInterval   = fmt.Errorf("Error: pollInterval must be positive\n")
	ErrInvalidReportInterval = fmt.Errorf("Error: reportInterval must be positive\n")
)
