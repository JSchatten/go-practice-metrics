package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/go-resty/resty/v2"
)

func sendMetricsJson(serverAddr string, memStorage *storage.MemStorage) error {
	client := resty.New()

	for _, metric := range memStorage.Metrics {
		// Проверяем, что метрика имеет хотя бы одно из значений
		if metric.Delta == nil && metric.Value == nil {
			return fmt.Errorf("metric %s has no value or delta", metric.ID)
		}

		// Определяем тип метрики
		metricType := metric.MType
		switch metricType {
		case MetricsModel.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("counter metric %s has no delta", metric.ID)
			}
		case MetricsModel.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("gauge metric %s has no value", metric.ID)
			}
		default:
			return fmt.Errorf("unsupported metric type: %s", metricType)
		}

		requestMetric := Metrics{
			ID:    metric.ID,
			MType: string(metric.MType),
			Delta: metric.Delta,
			Value: metric.Value,
		}

		// Отправляем POST-запрос
		// Сериализуем в JSON
		jsonData, err := json.Marshal(requestMetric)
		if err != nil {
			return fmt.Errorf("failed to marshal metrics: %w", err)
		}

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(jsonData).
			Post(fmt.Sprintf("http://%s/update/", serverAddr))

		if err != nil {
			return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("unexpected status code for metric %s: %d", metric.ID, resp.StatusCode())
		}
	}
	return nil
}
