package handler

import (
	"fmt"
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Method not allowed"})
			// badRequest(c)
			return
		}

		metricType := c.Param("type")
		metricName := c.Param("name")
		if metricType == "" || metricName == "" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Metric type and name is required"})
			// methodNotAllowed(c)
			return
		}

		metric := storage.GetMetric(metricName)

		if metric == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Metric not found"})
			// metricNotFound(c)
			return
		}

		var valueStr string
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Value is missing for gauge"})
				// valueNotProvided(c)
				return
			}
			valueStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		case model.Counter:
			if metric.Delta == nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delta is missing for counter"})
				// deltaNotProvided(c)
				return
			}
			valueStr = strconv.FormatInt(*metric.Delta, 10)
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unknown metric type: %s", metric.MType)})
			// unknownMetricType(c, metricType)
			return
		}

		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.WriteString(valueStr)
		c.AbortWithStatus(http.StatusOK)
	}
}
