package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

type Storage interface {
	UpdateMetric(metric *MetricsModel.Metrics) error
}

// Обработчик HTTP-запросов
func updateHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
			return
		}

		// Парсинг пути
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 5 || pathParts[1] != "update" {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
		// Тут рипаем по кускам
		metricType := pathParts[2]
		metricName := pathParts[3]
		valueStr := pathParts[4]

		// И проверяем на пустоту
		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		var metric *MetricsModel.Metrics

		switch metricType {
		case MetricsModel.Gauge:
			valueFloat, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid value format: %v", err), http.StatusBadRequest)
				return
			}
			metric = &MetricsModel.Metrics{
				ID:    metricName,
				MType: MetricsModel.Gauge,
				Value: &valueFloat,
			}
		case MetricsModel.Counter:
			valueInt, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid value format: %v", err), http.StatusBadRequest)
				return
			}
			metric = &MetricsModel.Metrics{
				ID:    metricName,
				MType: MetricsModel.Counter,
				Delta: &valueInt,
			}
		default:
			http.Error(w, fmt.Sprintf("Unknown metric type: %s", metricType), http.StatusBadRequest)
			return
		}

		// Обновление метрики
		if err := storage.UpdateMetric(metric); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %v", err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func liveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It's alive!"))
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	storageObj := storage.NewMemStorage()
	http.HandleFunc("/update/", updateHandler(storageObj))
	http.HandleFunc("/live/", liveHandler())
	http.HandleFunc("/", liveHandler())
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
