package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

// Обработчик для /value/<ТИП>/<ИМЯ>
func ValueHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			http.Error(w, "Method not allowed", http.StatusBadRequest)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 || pathParts[1] != "value" {
			http.Error(w, "Invalid URL format", http.StatusNotFound)
			return
		}

		fmt.Println("len(pathParts)", len(pathParts))

		metricType := pathParts[2]
		metricName := pathParts[3]

		// fmt.Println("pathParts", pathParts)
		// fmt.Println("metricType", metricType)
		// fmt.Println("metricName", metricName)

		if metricType == "" || metricName == "" {
			http.Error(w, "Metric type or name is required", http.StatusNotFound)
			return
		}

		metric := storage.GetMetric(metricName)

		if metric == nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		var valueStr string
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				http.Error(w, "Value is missing for gauge", http.StatusBadRequest)
				return
			}
			valueStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		case model.Counter:
			if metric.Delta == nil {
				http.Error(w, "Delta is missing for counter", http.StatusBadRequest)
				return
			}
			valueStr = strconv.FormatInt(*metric.Delta, 10)
		default:
			http.Error(w, fmt.Sprintf("Unknown metric type: %s", metric.MType), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(valueStr))
		w.WriteHeader(http.StatusOK)
	}
}
