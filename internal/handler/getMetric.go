package handler

import (
	"net/http"
	"strconv"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

// Обработчик для /value/<ТИП>/<ИМЯ>
func ValueHandler(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			badRequest(c)
			return
		}

		metricType := c.Param("type")
		metricName := c.Param("name")
		if metricType == "" || metricName == "" {
			methodNotAllowed(c)
			return
		}

		metric := storage.GetMetric(metricName)

		if metric == nil {
			metricNotFound(c)
			return
		}

		var valueStr string
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				valueNotProvided(c)
				return
			}
			valueStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		case model.Counter:
			if metric.Delta == nil {
				deltaNotProvided(c)
				return
			}
			valueStr = strconv.FormatInt(*metric.Delta, 10)
		default:
			unknownMetricType(c, metricType)
			return
		}

		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.WriteString(valueStr)
		c.AbortWithStatus(http.StatusOK)
	}
}
