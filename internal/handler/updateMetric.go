package handler

import (
	"net/http"
	"strconv"

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
			methodNotAllowed(c)
			return
		}

		metricType := c.Param("type")
		metricName := c.Param("name")
		valueStr := c.Param("value")
		if metricType == "" || metricName == "" || valueStr == "" {
			missingParameters(c)
			return
		}

		var metric *model.Metrics

		switch metricType {
		case model.Gauge:
			valueFloat, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				invalidValueFormat(c, err)
				return
			}
			metric = &model.Metrics{
				ID:    metricName,
				MType: model.Gauge,
				Value: &valueFloat,
			}
		case model.Counter:
			valueInt, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				invalidValueFormat(c, err)
				return
			}
			metric = &model.Metrics{
				ID:    metricName,
				MType: model.Counter,
				Delta: &valueInt,
			}
		default:
			unknownMetricType(c, metricType)
			return
		}

		// Обновление метрики
		if err := storage.UpdateMetric(metric); err != nil {
			failedToUpdateMetric(c, err)
			return
		}
		// fmt.Printf("Metrics %+v added \n", metric)
		c.JSON(http.StatusOK, gin.H{"status": "Metric updated"})
	}
}
