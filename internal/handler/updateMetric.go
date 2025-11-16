package handler

import (
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

// Обработчик HTTP-запросов
func UpdateHandler(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		logZero.Logger.Info().Msg("UpdateHandler")

		if c.Request.Method != http.MethodPost {
			MethodNotAllowed(c)
			return
		}

		metricType := c.Param("type")
		metricName := c.Param("name")
		valueStr := c.Param("value")
		if metricType == "" || metricName == "" || valueStr == "" {
			MissingParameters(c)
			return
		}

		metric, err := model.NewMetrics(metricName, metricType, valueStr)
		if err != nil {
			switch err {
			case model.ErrUnknownMetricType:
				UnknownMetricType(c, metricType)
			case model.ErrInvalidCounterValue, model.ErrInvalidGaugeValue:
				InvalidValueFormat(c, err)
			default:
				BadRequestVerbose(c, err)
			}
			return
		}

		// Обновление метрики
		if err := storage.UpdateMetric(c, metric); err != nil {
			FailedToUpdateMetric(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "Metric updated"})
	}
}
