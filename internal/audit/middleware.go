package audit

import (
	"time"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware возвращает Gin-мидлвару, которая:
//   - После выполнения обработчика
//   - Проверяет наличие данных аудита в контексте (`audit.metrics`)
//   - Если данные есть - формирует AuditEvent и отправляет в AuditManager
//
// Как использовать:
//
//  1. Обработчик должен добавить имена затронутых метрик в контекст:
//
//     c.Set("audit.metrics", []string{"poll_count", "requests"})
//
//  2. Мидлвара соберёт IP, время и вызовет Notify()
//
// Пример:
//
//	router.POST("/update", AuditMiddleware(manager), handler.UpdateHandler(storage))
//
// Зависимости:
//   - manager *AuditManager - должен быть предварительно сконфигурирован
//
// Поведение:
//   - Работает только после c.Next() - чтобы обработчик мог установить данные
//   - Игнорирует отсутствие или некорректные данные
//   - Не влияет на HTTP-ответ
func AuditMiddleware(manager *AuditManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // сначала дадим обработчику собрать данные

		// Передача после next будет через контекст, так лучше для gin
		// Обычно это делается через c.Set / c.Get
		metricsRaw, exists := c.Get("audit.metrics")
		if !exists {
			return
		}
		metrics, ok := metricsRaw.([]string)
		if !ok || len(metrics) == 0 {
			return
		}

		ip := c.ClientIP()

		event := AuditEvent{
			Timestamp: time.Now().Unix(),
			Metrics:   metrics,
			IPAddress: ip,
		}

		manager.Notify(event)
	}
}
