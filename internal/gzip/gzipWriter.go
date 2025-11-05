package middleware

import (
	"compress/gzip"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func newGzipWriter(w *gzip.Writer, rw gin.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: rw,
		writer:         w,
	}
}

// Write переопределяем для сжатия
func (w *gzipWriter) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}
