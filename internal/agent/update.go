package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
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
func sendMetrics(memStorage *storage.MemStorage) {
	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	fmt.Printf("Отправка метрик: %s\n", memStorage.String())
}

// Функция для обновления метрик из runtime
func UpdateRuntimeMetrics(storage *storage.MemStorage, done <-chan struct{}) {
	ticker_collect := time.NewTicker(pollInterval)
	ticker_send := time.NewTicker(reportInterval)
	defer ticker_collect.Stop()
	defer ticker_send.Stop()

	for {
		select {
		case <-ticker_send.C:
			sendMetrics(storage)
		case <-ticker_collect.C:
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
		case <-done:
			return
		}
	}
}
