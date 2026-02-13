package repository

import (
	"errors"
	"fmt"
)

var (
	ErrFilepathEmpty = errors.New("file path is not set")
	ErrFileRead      = errors.New("failed to read metrics file")
	ErrFileWrite     = errors.New("failed to write metrics file")
	ErrFileUnmarshal = errors.New("failed to unmarshal metrics from file")
	ErrFileMarshal   = errors.New("failed to marshal metrics")
	ErrFileNoData    = errors.New("no data provided for saving into file")
)

// Форматированные ошибки

func NewErrReadFile(path string, err error) error {
	return fmt.Errorf("%w '%s': %w", ErrFileRead, path, err)
}

func NewErrWriteFile(path string, err error) error {
	return fmt.Errorf("%w '%s': %w", ErrFileWrite, path, err)
}

func NewErrUnmarshal(path string, err error) error {
	return fmt.Errorf("%w '%s': %w", ErrFileUnmarshal, path, err)
}

func NewErrMarshal(err error) error {
	return fmt.Errorf("%w: %w", ErrFileMarshal, err)
}
