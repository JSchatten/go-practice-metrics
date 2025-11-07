package agent

import (
	"fmt"
	"net/http"
	"strconv"

	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/go-resty/resty/v2"
)

func sendMetrics(serverAdress string, memStorage *storage.MemStorage) error {
	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	// fmt.Printf("Отправка метрик: %s\n", memStorage.String())
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

		// Формируем URL в зависимости от типа метрики
		var valueStr string
		switch metricType {
		case MetricsModel.Counter:
			valueStr = strconv.FormatInt(*metric.Delta, 10)
		case MetricsModel.Gauge:
			valueStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		}

		url := fmt.Sprintf("http://%s/update/%s/%s/%s", serverAdress, metricType, metric.ID, valueStr)

		// Отправляем POST-запрос
		resp, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)

		if err != nil {
			return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("unexpected status code for metric %s: %d", metric.ID, resp.StatusCode())
		}
	}
	return nil
}
