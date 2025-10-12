package handler

import (
	"fmt"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storageService "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

// Для проверки
func LiveHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"status": "It's alive!"})
	}
}

// Обработчик для корневого пути /
func RootHandler(storage storageService.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			// http.Error(w, "Method not allowed", http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Method not allowed"})
			return
		}

		// Получаем все метрики
		metrics := make([]model.Metrics, 0, len(storage.(*storageService.MemStorage).Metrics))
		for _, metric := range storage.(*storageService.MemStorage).Metrics {
			metrics = append(metrics, *metric)
		}

		// Генерируем HTML
		html := "<html><body><h1>Metrics</h1><table border='1'><tr><th>ID</th><th>Type</th><th>Value/Delta</th></tr>"
		for _, metric := range metrics {
			var value string
			switch metric.MType {
			case model.Gauge:
				if metric.Value == nil {
					value = "N/A"
				} else {
					value = fmt.Sprintf("%.2f", *metric.Value)
				}
			case model.Counter:
				if metric.Delta == nil {
					value = "N/A"
				} else {
					value = fmt.Sprintf("%d", *metric.Delta)
				}
			default:
				value = "Unknown type"
			}
			html += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", metric.ID, metric.MType, value)
		}
		html += "</table></body></html>"

		// w.Header().Set("Content-Type", "text/html")
		// w.Write([]byte(html))
		// w.WriteHeader(http.StatusOK)
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}
