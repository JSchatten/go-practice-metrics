package config

import (
	"errors"
)

var (
	ErrInvalidPollInterval   = errors.New("error: pollInterval must be positive")
	ErrInvalidReportInterval = errors.New("error: reportInterval must be positive")
	ErrInvalidStoreInterval  = errors.New("error: storeInterval must be positive or equal zero")
	ErrInvalidDSN            = errors.New("DATABASE_DSN is provided but empty")
)
