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

		// fmt.Println(r.URL.Path)
		if c.Request.Method != http.MethodPost {
			// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			// http.Error(w, "Method not allowed", http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Method not allowed"})
			return
		}

		// if r.Header.Get("Content-Type") != "text/plain" {
		// 	fmt.Println("Invalid Content-Type")
		// 	http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		// 	return
		// }

		// Парсинг пути
		// pathParts := strings.Split(r.URL.Path, "/")
		// if len(pathParts) < 5 || pathParts[1] != "update" {
		// 	// http.Error(w, "Invalid URL format", http.StatusBadRequest)
		// 	// http.Error(w, "Invalid URL format", http.StatusNotFound)
		// 	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Missing required parameters"})
		// 	return
		// }

		// Тут рипаем по кускам
		metricType := c.Param("type")
		metricName := c.Param("name")
		valueStr := c.Param("value")
		if metricType == "" || metricName == "" || valueStr == "" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Missing required parameters"})
			return
		}
		// fmt.Println("metricType", metricType)
		// fmt.Println("metricName", metricName)
		// fmt.Println("valueStr", valueStr)

		var metric *model.Metrics

		switch metricType {
		case model.Gauge:
			valueFloat, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				// http.Error(w, fmt.Sprintf("Invalid value format: %v", err), http.StatusBadRequest)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid value format: %v", err)})
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
				// http.Error(w, fmt.Sprintf("Invalid value format: %v", err), http.StatusBadRequest)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
				return
			}
			metric = &model.Metrics{
				ID:    metricName,
				MType: model.Counter,
				Delta: &valueInt,
			}
		default:
			// http.Error(w, fmt.Sprintf("Unknown metric type: %s", metricType), http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unknown metric type: %s", metricType)})
			return
		}

		// Обновление метрики
		if err := storage.UpdateMetric(metric); err != nil {
			// http.Error(w, fmt.Sprintf("Failed to update metric: %v", err), http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to update metric: %v", err)})
			return
		}

		// fmt.Printf("Metrics %+v added \n", metric)

		// w.WriteHeader(http.StatusOK)
		c.JSON(http.StatusOK, gin.H{"status": "Metric updated"})
	}
}
