// server.go - сервер для обработки запросов, предоставление API
package handler

import (
	"fmt"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storageService "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

const (
	htmlTop    = "<html><body><h1>Metrics</h1><table border='1'><tr><th>ID</th><th>Type</th><th>Value/Delta</th></tr>"
	htmlBottom = "</table></body></html>"
)

// LiveHandler возвращает статус "It's alive!" для проверки доступности сервера.
//
// Метод: любой (но всегда возвращает 405 - Method Not Allowed)
// Путь: /live
//
// Примечание: эндпоинт **намеренно возвращает 405**, чтобы продемонстрировать обработку ошибок.
// В реальном приложении следует использовать 200 OK.
//
// BUG(@JSchatten): Тестовый баг для проверки работы генерации godoc:
// Допустим неверное использование ручки, что может привести к некорректной работе.
func LiveHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"status": "It's alive!"})
	}
}

// PingDatabaseHandler проверяет соединение с базой данных.
//
// Метод: GET
// Путь: /ping
//
// В случае успеха:
//   - Код 200 OK
//   - Ответ: {"status": "ok"}
//
// В случае ошибки:
//   - Код 500 Internal Server Error
//   - Ответ: {"error": "cannot connect to database"}
//
// Логирование:
//   - Ошибки подключения логируются как Error.
//
// Зависимости:
//   - storage.Storage - должен реализовывать метод PingDatabase
func PingDatabaseHandler(storage storageService.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := storage.PingDatabase(c); err != nil {
			logZero.Logger.Error().Err(err).Msg("DB ping failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot connect to database"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// RootHandler возвращает HTML-страницу со списком всех метрик.
//
// Метод: GET
// Путь: /
//
// Генерирует простую HTML-таблицу со всеми сохранёнными метриками.
//
// Формат:
//   - Показывает ID, тип и значение (или "N/A", если значение nil)
//   - Counter: отображается как целое число
//   - Gauge: отображается с двумя знаками после запятой
//
// Ошибки:
//   - 405 Method Not Allowed, если метод не GET
func RootHandler(storage storageService.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			MethodNotAllowed(c)
			return
		}

		// Получаем все метрики
		metrics := make([]model.Metrics, 0, len(storage.(*storageService.MemStorage).Metrics))
		for _, metric := range storage.(*storageService.MemStorage).Metrics {
			metrics = append(metrics, *metric)
		}

		// Генерируем HTML
		html := htmlTop
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
		html += htmlBottom

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}
