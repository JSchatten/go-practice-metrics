package handler

import (
	"encoding/json"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

func UpdateHandlerJSON(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		logZero.Logger.Info().Msg("UpdateHandlerJSON")
		var metricIn model.RequestMetrics

		if c.Request.Method != http.MethodPost {
			methodNotAllowed(c)
			return
		}

		// Это по-хорошему, т.к. есть sonic-avx, для build-тегов, но требования такие
		// if err := c.ShouldBindJSON(&metricIn); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		// 	return
		// }

		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&metricIn); err != nil {
			bodyInvalidJSON(c)
			return
		}

		if metricIn.ID == "" || metricIn.MType == "" {
			bodyMissingFields(c)
			return
		}

		switch metricIn.MType {
		case "counter":
			if metricIn.Delta == nil {
				deltaNotProvided(c)
				return
			}
		case "gauge":
			if metricIn.Value == nil {
				valueNotProvided(c)
				return
			}
		default:
			bodyInvalidMetricType(c)
			return
		}

		metricOut := model.Metrics{
			ID:    metricIn.ID,
			MType: metricIn.MType,
			Delta: metricIn.Delta,
			Value: metricIn.Value,
		}

		err := storage.UpdateMetric(&metricOut)
		if err != nil {
			failedToUpdateMetric(c, err)
			return
		}

		c.JSON(http.StatusOK, metricIn)
	}
}
