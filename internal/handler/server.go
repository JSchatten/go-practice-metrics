package handler

import (
	"fmt"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storageService "github.com/JSchatten/go-practice-metrics/internal/service"
)

// Для проверки
func LiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It's alive!"))
		w.WriteHeader(http.StatusOK)
	}
}

// Обработчик для корневого пути /
func RootHandler(storage storageService.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			http.Error(w, "Method not allowed", http.StatusBadRequest)
			return
		}

		// Получаем все метрики
		metrics := make([]model.Metrics, 0, len(storage.(*storageService.MemStorage).Metrics))
		for _, metric := range storage.(*storageService.MemStorage).Metrics {
			metrics = append(metrics, *metric)
		}

		// Генерируем HTML
		html := "<html><body><h1>Metrics</h1><table border='1'><tr><th>ID</th><th>Type</th><th>Value/Delta</th></tr>"
		for _, metric := range metrics {
			var value string
			switch metric.MType {
			case model.Gauge:
				if metric.Value == nil {
					value = "N/A"
				} else {
					value = fmt.Sprintf("%.2f", *metric.Value)
				}
			case model.Counter:
				if metric.Delta == nil {
					value = "N/A"
				} else {
					value = fmt.Sprintf("%d", *metric.Delta)
				}
			default:
				value = "Unknown type"
			}
			html += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", metric.ID, metric.MType, value)
		}
		html += "</table></body></html>"

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		w.WriteHeader(http.StatusOK)
	}
}
