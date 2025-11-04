// internal/middleware/gzip.go
package middleware

import (
	"bufio"
	"compress/gzip"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipWriter — обёртка над gin.ResponseWriter
type gzipWriter struct {
	writer   *gzip.Writer
	response gin.ResponseWriter
}

func newGzipWriter(w *gzip.Writer, rw gin.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		writer:   w,
		response: rw,
	}
}

func (w *gzipWriter) Status() int {
	return w.Status()
}

func (w *gzipWriter) WriteHeaderNow() {
	w.WriteHeaderNow()
}

func (w *gzipWriter) WriteString(s string) (int, error) {
	return w.WriteString(s)
}

func (w *gzipWriter) Size() int {
	return w.Size()
}

func (w *gzipWriter) Header() http.Header {
	return w.response.Header()
}

func (w *gzipWriter) Pusher() http.Pusher {
	return w.response.Pusher()
}

func (w *gzipWriter) Written() bool {
	return w.Written()
}

func (w *gzipWriter) CloseNotify() <-chan bool {
	return w.response.(http.CloseNotifier).CloseNotify()
}

func (w *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.response.(http.Hijacker).Hijack()
}

// Flush вызывает сначала gzip.Flush, затем underlying Flush
func (w *gzipWriter) Flush() {
	// 	if !w.wroteHeader {
	// 		w.WriteHeader(http.StatusOK)
	// 	}
	// 	w.writer.Flush()
	// 	if fw, ok := w.response.(http.Flusher); ok {
	// 		fw.Flush()
	// 	}
	w.writer.Flush() // сбрасывает буфер gzip
	w.response.Flush()
}

func (w *gzipWriter) Write(data []byte) (int, error) {
	// Проверяем Content-Type
	contentType := w.response.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") &&
		!strings.HasPrefix(contentType, "text/html") {
		// Отключаем сжатие
		w.response.Header().Del("Content-Encoding")
		w.response.Header().Del("Vary")
		return w.response.Write(data)
	}
	return w.writer.Write(data)
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	// Аналогично — проверяем тип
	contentType := w.response.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") &&
		!strings.HasPrefix(contentType, "text/html") {
		w.response.Header().Del("Content-Encoding")
		w.response.Header().Del("Vary")
	}
	w.response.WriteHeader(statusCode)
}

// GzipMiddleware возвращает Gin middleware для обработки сжатия
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		contentEncoding := c.Request.Header.Get("Content-Encoding")
		isGzipped := contentEncoding == "gzip"

		// if isGzipped {
		// 	reader, err := gzip.NewReader(c.Request.Body)
		// 	if err != nil {
		// 		c.AbortWithStatus(http.StatusBadRequest)
		// 		return
		// 	}
		// 	defer reader.Close()
		// 	c.Request.Body = reader
		// }

		// if supportsGzip {
		// 	gz := gzip.NewWriter(c.Writer)
		// 	c.Writer = &gzipWriter{
		// 		writer:   gz,
		// 		response: c.Writer,
		// 	}

		// 	// Устанавливаем заголовки
		// 	c.Header("Content-Encoding", "gzip")
		// 	c.Header("Vary", "Accept-Encoding")

		// 	defer gz.Close()
		// }

		// Распаковка тела, если сжато
		if isGzipped {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer reader.Close()
			c.Request.Body = reader
		}

		// Если клиент поддерживает gzip — оборачиваем Writer
		if supportsGzip {
			gz := gzip.NewWriter(c.Writer)
			gw := newGzipWriter(gz, c.Writer)

			c.Writer = gw
			defer gz.Close() // важен порядок: после завершения запроса
		}

		c.Next()

	}
}
