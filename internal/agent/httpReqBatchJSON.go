package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	gzip "github.com/JSchatten/go-practice-metrics/internal/gzip"
	hashprocess "github.com/JSchatten/go-practice-metrics/internal/hashprocess"
	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/go-resty/resty/v2"
)

func sendMetricsBatchJSON(client *resty.Client, serverAddr string, memStorage *storage.MemStorage, HashKey string) error {
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

	compressed, err := gzip.CompressGZIP(jsonData)
	if err != nil {
		return errors.New("failed to compress metrics")
	}

	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip")

	if HashKey != "" {
		signOfRequest := hashprocess.Sign(compressed, HashKey)
		request = request.SetHeader("HashSHA256", signOfRequest)
	}

	request = request.SetBody(compressed)

	var lastErr error
	var delay time.Duration

	// TODO Это во флаги по-хорошему, но кто знает. что будет дальше
	const RetryTimeoutDelta = 2
	const MaxRetries = 5
	const AttemptCount = 5

	for attempt := range AttemptCount {
		if attempt > 0 {
			fmt.Printf("Retry %d/%d in %v...\n", attempt, MaxRetries, delay)
			time.Sleep(delay)
			delay += RetryTimeoutDelta * time.Second // Линейное увеличение
		}

		resp, err := request.Post(fmt.Sprintf("http://%s/updates/", serverAddr))

		if err == nil && resp.StatusCode() == http.StatusOK {
			// По-хорошему бы проверять хэш от сервера
			// fmt.Println("resp hash: ", resp.Header().Get("HashSHA256"))
			fmt.Println("Metrics sent successfully")
			return nil
		}

		lastErr = fmt.Errorf("send failed: status=%d, err=%w", resp.StatusCode(), err)
		fmt.Printf("Send attempt %d failed: %v\n", attempt, lastErr)
	}

	return lastErr
}
