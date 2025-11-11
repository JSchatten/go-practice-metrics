package config

import (
	"fmt"
)

var (
	ErrInvalidPollInterval   = fmt.Errorf("error: pollInterval must be positive")
	ErrInvalidReportInterval = fmt.Errorf("error: reportInterval must be positive")
	ErrInvalidStoreInterval  = fmt.Errorf("error: storeInterval must be positive or equal zero")
	ErrInvalidDSN            = fmt.Errorf("DATABASE_DSN is provided but empty")
)
