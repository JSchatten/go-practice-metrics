package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"

	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/go-resty/resty/v2"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverAddr     = "127.0.0.1:8080"
)

func getMetricGauge(id string, value float64) *MetricsModel.Metrics {
	return &MetricsModel.Metrics{
		ID:    id,
		MType: MetricsModel.Gauge,
		Value: &value,
	}
}

func getMetricCount(id string, delta int64) *MetricsModel.Metrics {
	return &MetricsModel.Metrics{
		ID:    id,
		MType: MetricsModel.Gauge,
		Delta: &delta,
	}
}

// Функция-заглушка для отправки метрик
func sendMetrics(memStorage *storage.MemStorage) error {
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

		url := fmt.Sprintf("http://%s/update/%s/%s/%s", serverAddr, metricType, metric.ID, valueStr)

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

// Функция для обновления метрик из runtime
func UpdateRuntimeMetrics(storage *storage.MemStorage, done <-chan struct{}) {
	tickerCollect := time.NewTicker(pollInterval)
	tickerSend := time.NewTicker(reportInterval)
	defer tickerCollect.Stop()
	defer tickerSend.Stop()

	for {
		select {
		case <-tickerSend.C:
			fmt.Println("Sending metrics...")
			err := sendMetrics(storage)
			if err != nil {
				fmt.Printf("Error sending metrics: %v\n", err)
			} else {
				fmt.Println("Sended successful")
			}
		case <-tickerCollect.C:
			fmt.Println("Collecting metrics...")
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			// Большой список, 1e9 для перевода в секунды
			storage.UpdateMetric(getMetricGauge("Alloc", float64(memStats.Alloc)))
			storage.UpdateMetric(getMetricGauge("BuckHashSys", float64(memStats.BuckHashSys)))
			storage.UpdateMetric(getMetricGauge("Frees", float64(memStats.Frees)))
			storage.UpdateMetric(getMetricGauge("GCCPUFraction", memStats.GCCPUFraction))
			storage.UpdateMetric(getMetricGauge("GCSys", float64(memStats.GCSys)))
			storage.UpdateMetric(getMetricGauge("HeapAlloc", float64(memStats.HeapAlloc)))
			storage.UpdateMetric(getMetricGauge("HeapIdle", float64(memStats.HeapIdle)))
			storage.UpdateMetric(getMetricGauge("HeapInuse", float64(memStats.HeapInuse)))
			storage.UpdateMetric(getMetricGauge("HeapObjects", float64(memStats.HeapObjects)))
			storage.UpdateMetric(getMetricGauge("HeapReleased", float64(memStats.HeapReleased)))
			storage.UpdateMetric(getMetricGauge("HeapSys", float64(memStats.HeapSys)))
			storage.UpdateMetric(getMetricGauge("LastGC", float64(memStats.LastGC)/1e9))
			storage.UpdateMetric(getMetricGauge("Lookups", float64(memStats.Lookups)))
			storage.UpdateMetric(getMetricGauge("MCacheInuse", float64(memStats.MCacheInuse)))
			storage.UpdateMetric(getMetricGauge("MCacheSys", float64(memStats.MCacheSys)))
			storage.UpdateMetric(getMetricGauge("MSpanInuse", float64(memStats.MSpanInuse)))
			storage.UpdateMetric(getMetricGauge("MSpanSys", float64(memStats.MSpanSys)))
			storage.UpdateMetric(getMetricGauge("Mallocs", float64(memStats.Mallocs)))
			storage.UpdateMetric(getMetricGauge("NextGC", float64(memStats.NextGC)))
			storage.UpdateMetric(getMetricGauge("NumForcedGC", float64(memStats.NumForcedGC)))
			storage.UpdateMetric(getMetricGauge("NumGC", float64(memStats.NumGC)))
			storage.UpdateMetric(getMetricGauge("OtherSys", float64(memStats.OtherSys)))
			storage.UpdateMetric(getMetricGauge("PauseTotalNs", float64(memStats.PauseTotalNs)/1e9))
			storage.UpdateMetric(getMetricGauge("StackInuse", float64(memStats.StackInuse)))
			storage.UpdateMetric(getMetricGauge("StackSys", float64(memStats.StackSys)))
			storage.UpdateMetric(getMetricGauge("Sys", float64(memStats.Sys)))
			storage.UpdateMetric(getMetricGauge("TotalAlloc", float64(memStats.TotalAlloc)))
			// Дополнительные
			storage.UpdateMetric(getMetricGauge("RandomValue", float64(rand.IntN(100))))
			pollCnt := storage.GetMetric("PollCount")
			if pollCnt == nil {
				storage.UpdateMetric(getMetricCount("PollCount", 1))
			} else {
				storage.UpdateMetric(getMetricCount("PollCount", *pollCnt.Delta+1))
			}
			fmt.Println("Collecting metrics done")
		case <-done:
			fmt.Println("Stop processing metrics")
			return
		}
	}
}
