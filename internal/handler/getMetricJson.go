package handler

import (
	"encoding/json"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

func ValueHandlerJSON(storage storage.Storage) gin.HandlerFunc {
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
		logZero.Logger.Info().Msgf("UpdateHandlerJSON metrics in = %s", metric.String())

	}
}
