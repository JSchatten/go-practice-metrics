package handler

import (
	"encoding/json"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

func UpdateHandlerJson(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// На случай тестов
		// contentType := c.Request.Header.Get("Content-Type")
		// if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "content-type must be application/json"})
		// 	return
		// }

		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&metricIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
			return
		}

		if metricIn.ID == "" || metricIn.MType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: id or type"})
			return
		}

		switch metricIn.MType {
		case "counter":
			if metricIn.Delta == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing delta for counter"})
				return
			}
		case "gauge":
			if metricIn.Value == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing value for gauge"})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric type: must be 'gauge' or 'counter'"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update metric"})
			return
		}

		c.JSON(http.StatusOK, metricIn)
	}
}
