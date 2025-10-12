package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

// Для проверки
func LiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It's alive!"))
		w.WriteHeader(http.StatusOK)
	}
}

// Обработчик HTTP-запросов
func UpdateHandler(storage storage.Storage) http.HandlerFunc {
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

		// fmt.Println(metricType)
		// fmt.Println(metricName)
		// fmt.Println(valueStr)

		// И проверяем на пустоту
		if metricType == "" {
			http.Error(w, "Metric type name is required", http.StatusNotFound)
			return
		}
		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}
		if valueStr == "" {
			http.Error(w, "Metric Value is required", http.StatusNotFound)
			return
		}

		var metric *model.Metrics

		switch metricType {
		case model.Gauge:
			valueFloat, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid value format: %v", err), http.StatusBadRequest)
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
				http.Error(w, fmt.Sprintf("Invalid value format: %v", err), http.StatusBadRequest)
				return
			}
			metric = &model.Metrics{
				ID:    metricName,
				MType: model.Counter,
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
