package repository

import "fmt"

var (
	ErrFilePathEmpty = fmt.Errorf("file path is not set")
	ErrReadFile      = fmt.Errorf("failed to read metrics file")
	ErrWriteFile     = fmt.Errorf("failed to write metrics file")
	ErrUnmarshal     = fmt.Errorf("failed to unmarshal metrics from file")
	ErrMarshal       = fmt.Errorf("failed to marshal metrics")
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
