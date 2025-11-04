package handler

import (
	"encoding/json"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

func ValueHandlerJSON(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var metricIn model.RequestMetrics

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
			// c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
			bodyInvalidJSON(c)
			return
		}

		if metricIn.ID == "" || metricIn.MType == "" {
			// c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: id or type"})
			bodyMissingFields(c)
			return
		}

		switch metricIn.MType {
		case "gauge", "counter":
			// ок
		default:
			// c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type: must be 'gauge' or 'counter'"})
			bodyInvalidMetricType(c)
			return
		}

		metric := storage.GetMetric(metricIn.ID)

		if metric.MType != metricIn.MType {
			// c.JSON(http.StatusBadRequest, gin.H{"error": "Metric type mismatch"})
			invalidMetricTypeMissmatch(c)
			return
		}

		c.JSON(http.StatusOK, metric)

	}
}
