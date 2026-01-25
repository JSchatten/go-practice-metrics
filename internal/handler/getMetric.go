package handler

import (
	"encoding/json"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

// ValueHandler возвращает значение метрики по её имени и типу в формате JSON.
//
// Метод: POST
// Путь: /value
//
// Ожидает JSON-тело с полями:
//   - id (string): имя метрики
//   - type (string): тип метрики — "gauge" или "counter"
//
// Пример запроса:
//
//	POST /value HTTP/1.1
//	Content-Type: application/json
//
//	{
//	  "id": "poll_count",
//	  "type": "counter"
//	}
//
// В случае успеха возвращает:
//   - Код 200 OK
//   - JSON-объект с полным описанием метрики (включая значение)
//
// Пример ответа:
//
//	{
//	  "id": "poll_count",
//	  "type": "counter",
//	  "delta": 42
//	}
//
// Ошибки:
//   - 400 Bad Request: невалидный JSON, отсутствуют поля `id` или `type`
//   - 404 Not Found: метрика с таким именем не найдена
//   - 400 Bad Request: тип метрики не "gauge" и не "counter"
//   - 409 Conflict: запрошенный тип не совпадает с типом сохранённой метрики
//
// Логирование:
//   - Запрос логируется на уровне Info с указанием входной и выходной метрики.
//
// Зависимости:
//   - storage.Storage — интерфейс хранилища, реализующий метод GetMetric
//
// Примечание: рекомендуется использовать этот эндпоинт вместо устаревшего GET /value/{type}/{name}.
func ValueHandler(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		logZero.Logger.Info().Msg("ValueHandlerJSON")
		var metricIn model.Metrics

		// Это по-хорошему, т.к. есть sonic-avx, для build-тегов, но требования такие
		// if err := c.ShouldBindJSON(&metricIn); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		// 	return
		// }

		if err := json.NewDecoder(c.Request.Body).Decode(&metricIn); err != nil {
			BodyInvalidJSON(c)
			return
		}

		if metricIn.ID == "" || metricIn.MType == "" {
			BodyMissingFields(c)
			return
		}

		switch metricIn.MType {
		case "gauge", "counter":
			// ок
		default:
			BodyInvalidMetricType(c)
			return
		}
		logZero.Logger.Info().Msgf("ValueHandlerJSON metrics in = %s", metricIn.String())

		metric := storage.GetMetric(c, metricIn.ID)
		if metric == nil {
			MetricNotFound(c)
			return
		}

		if metric.MType != metricIn.MType {
			InvalidMetricTypeMismatch(c)
			return
		}

		c.JSON(http.StatusOK, metric)
		logZero.Logger.Info().Msgf("ValueHandlerJSON metrics out = %s", metric.String())

	}
}
