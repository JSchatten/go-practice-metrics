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
			// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			// http.Error(w, "Method not allowed", http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Method not allowed"})
			return
		}

		// pathParts := strings.Split(r.URL.Path, "/")
		// if len(pathParts) < 3 || pathParts[1] != "value" {
		// 	// http.Error(w, "Invalid URL format", http.StatusNotFound)
		// 	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not found"})
		// 	return
		// }

		// fmt.Println("len(pathParts)", len(pathParts))

		// metricType := pathParts[2]
		// metricName := pathParts[3]

		metricType := c.Param("type")
		metricName := c.Param("name")

		// fmt.Println("pathParts", pathParts)
		// fmt.Println("metricType", metricType)
		// fmt.Println("metricName", metricName)

		if metricType == "" || metricName == "" {
			// http.Error(w, "Metric type or name is required", http.StatusNotFound)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Metric type and name is required"})
			return
		}

		metric := storage.GetMetric(metricName)

		if metric == nil {
			// http.Error(w, "Metric not found", http.StatusNotFound)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Metric not found"})
			return
		}

		var valueStr string
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				// http.Error(w, "Value is missing for gauge", http.StatusBadRequest)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Value is missing for gauge"})
				return
			}
			valueStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		case model.Counter:
			if metric.Delta == nil {
				// http.Error(w, "Delta is missing for counter", http.StatusBadRequest)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delta is missing for counter"})
				return
			}
			valueStr = strconv.FormatInt(*metric.Delta, 10)
		default:
			// http.Error(w, fmt.Sprintf("Unknown metric type: %s", metric.MType), http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unknown metric type: %s", metric.MType)})
			return
		}

		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.WriteString(valueStr)
		c.AbortWithStatus(http.StatusOK)
	}
}
