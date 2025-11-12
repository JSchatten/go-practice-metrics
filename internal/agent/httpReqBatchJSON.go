package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/go-resty/resty/v2"
)

func sendMetricsBatchJSON(serverAddr string, memStorage *storage.MemStorage) error {
	client := resty.New()

	var sendingMetrics []MetricsModel.Metrics

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

		sendingMetrics = append(sendingMetrics, MetricsModel.Metrics{
			ID:    metric.ID,
			MType: string(metric.MType),
			Delta: metric.Delta,
			Value: metric.Value,
		})

	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(sendingMetrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	compressed, err := CompressGZIP(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress metrics")
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		// SetBody(jsonData).
		SetBody(compressed).
		Post(fmt.Sprintf("http://%s/updates/", serverAddr))

	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code for metrics: %d", resp.StatusCode())
	}

	return nil
}
