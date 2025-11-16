package handler

import (
	"net/http"
	"strconv"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

// Обработчик для /value/<ТИП>/<ИМЯ>
func ValueHandler(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		logZero.Logger.Info().Msg("ValueHandler")
		if c.Request.Method != http.MethodGet {
			BadRequest(c)
			return
		}

		metricType := c.Param("type")
		metricName := c.Param("name")
		if metricType == "" || metricName == "" {
			MethodNotAllowed(c)
			return
		}

		metric := storage.GetMetric(c, metricName)

		if metric == nil {
			MetricNotFound(c)
			return
		}

		var valueStr string
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				ValueNotProvided(c)
				return
			}
			valueStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		case model.Counter:
			if metric.Delta == nil {
				DeltaNotProvided(c)
				return
			}
			valueStr = strconv.FormatInt(*metric.Delta, 10)
		default:
			UnknownMetricType(c, metricType)
			return
		}

		c.Writer.Header().Set("Content-Type", "text/plain")
		_, err := c.Writer.WriteString(valueStr)
		if err != nil {
			logZero.Logger.Fatal().
				Err(err).
				Msg("Failed to write response")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.AbortWithStatus(http.StatusOK)
	}
}
