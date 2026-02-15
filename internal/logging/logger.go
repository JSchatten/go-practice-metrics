// Package logging предоставляет middleware для логирования HTTP-запросов и ответов
// в формате, совместимом с Gin-фреймворком.
//
// Основные функции:
//   - Создание кастомного ResponseWriter для сбора информации о статусе и размере ответа.
//   - Middleware для логирования входящих запросов и исходящих ответов.
package logging

import "github.com/gin-gonic/gin"

type responseWriter struct {
	gin.ResponseWriter
	status   int
	bodySize int
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.bodySize += size
	return size, err
}

func (rw *responseWriter) BodySize() int {
	return rw.bodySize
}
