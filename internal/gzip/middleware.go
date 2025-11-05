package gzip

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipMiddleware возвращает Gin middleware для обработки сжатия
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gz.Close()
			c.Request.Body = gz
		}

		if !acceptsGzip(c.Request) {
			c.Next()
			return
		}

		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()
		gw := newGzipWriter(gz, c.Writer)
		c.Writer = gw

		contentType := c.Writer.Header().Get("Content-Type")
		if shouldCompressContentType(contentType) {
			c.Writer.Header().Set("Content-Encoding", "gzip")
			c.Writer.Header().Set("Vary", "Accept-Encoding")
			// Удаляем размер, так как он будет неточный из-за сжатия
			c.Writer.Header().Del("Content-Length")
		} else {
			// Удаляем сжатие для бинарных типов, явно
			c.Writer.Header().Del("Content-Encoding")
			c.Writer.Header().Del("Vary")
		}
		c.Next()
	}
}

// acceptsGzip проверяет, что клиент поддерживает gzip
func acceptsGzip(r *http.Request) bool {
	accept := r.Header.Get("Accept-Encoding")
	parts := strings.Split(accept, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "gzip" || strings.HasPrefix(part, "gzip;") {
			return true
		}
	}
	return false
}

// shouldCompressContentType определяет, стоит ли сжимать контент
func shouldCompressContentType(contentType string) bool {
	if contentType == "" {
		return true // по умолчанию сжимаем
	}
	return strings.HasPrefix(contentType, "text/") ||
		strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "application/javascript")
}
