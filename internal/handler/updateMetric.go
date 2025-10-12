package handler

import (
	"fmt"
	"net/http"
	"strconv"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

// Обработчик HTTP-запросов
func UpdateHandler(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Method not allowed"})
			// methodNotAllowed(c)
			return
		}

		metricType := c.Param("type")
		metricName := c.Param("name")
		valueStr := c.Param("value")
		if metricType == "" || metricName == "" || valueStr == "" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Missing required parameters"})
			// missingParameters(c)
			return
		}

		var metric *model.Metrics

		switch metricType {
		case model.Gauge:
			valueFloat, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid value format: %v", err)})
				// invalidValueFormat(c, err)
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
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
				// invalidValueFormat(c, err)
				return
			}
			metric = &model.Metrics{
				ID:    metricName,
				MType: model.Counter,
				Delta: &valueInt,
			}
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unknown metric type: %s", metricType)})
			// unknownMetricType(c, metricType)
			return
		}

		// Обновление метрики
		if err := storage.UpdateMetric(metric); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to update metric: %v", err)})
			// failedToUpdateMetric(c, err)
			return
		}
		// fmt.Printf("Metrics %+v added \n", metric)
		c.JSON(http.StatusOK, gin.H{"status": "Metric updated"})
	}
}
