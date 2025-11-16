package repository

import (
	"errors"
	"fmt"
)

var (
	ErrFilePathEmpty = errors.New("file path is not set")
	ErrReadFile      = errors.New("failed to read metrics file")
	ErrWriteFile     = errors.New("failed to write metrics file")
	ErrUnmarshal     = errors.New("failed to unmarshal metrics from file")
	ErrMarshal       = errors.New("failed to marshal metrics")
	ErrNoData        = errors.New("no data provided for saving into file")
)

// Форматированные ошибки
func NewErrReadFile(path string, err error) error {
	return fmt.Errorf("%w '%s': %w", ErrReadFile, path, err)
}

func NewErrWriteFile(path string, err error) error {
	return fmt.Errorf("%w '%s': %w", ErrWriteFile, path, err)
}

func NewErrUnmarshal(path string, err error) error {
	return fmt.Errorf("%w '%s': %w", ErrUnmarshal, path, err)
}

func NewErrMarshal(err error) error {
	return fmt.Errorf("%w: %w", ErrMarshal, err)
}
