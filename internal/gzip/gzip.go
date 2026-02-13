// Package gzip предоставляет утилиты для сжатия и распаковки данных в формате GZIP.
//
// Основные функции:
//   - CompressGZIP: сжатие байтового массива с использованием GZIP.
//   - Middleware для автоматического сжатия ответов и распаковки тела запросов.
package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

func CompressGZIP(data []byte) ([]byte, error) {
	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	_, err := gz.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write gzip data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return compressed.Bytes(), nil
}
