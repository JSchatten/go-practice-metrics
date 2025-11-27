package config

import (
	"errors"
)

var (
	ErrInvalidPollInterval   = errors.New("error: pollInterval must be positive")
	ErrInvalidReportInterval = errors.New("error: reportInterval must be positive")
	ErrInvalidStoreInterval  = errors.New("error: storeInterval must be positive or equal zero")
	ErrInvalidDSN            = errors.New("error: DATABASE_DSN is provided but empty")
	ErrInvalidRateLimit      = errors.New("error: Rate limit must be positive")
	// ErrInvalidKey            = errors.New("KEY is provided but empty")
)
