package audit

import (
	"time"

	"github.com/gin-gonic/gin"
)

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
