package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/JSchatten/go-practice-metrics/internal/config"
	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

type flags struct {
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func InitFlags(
	pollInterval int,
	reportInterval int,
	serverAddr string,
) *flags {
	if pollInterval == 0 {
		pollInterval = 2
	}
	if reportInterval == 0 {
		reportInterval = 10
	}
	// Может подуммать над зависимостью интервала отправки от опроса
	// чтобы ограничить пользователя от безобразий
	// else if reportInterval < pollInterval {
	// reportInterval = pollInterval + 1
	// }
	return &flags{
		PollInterval:   time.Duration(pollInterval) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
		ServerAddr:     serverAddr,
	}
}

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
		MType: MetricsModel.Counter,
		Delta: &delta,
	}
}

// Функция для обновления метрик из runtime
func UpdateRuntimeMetrics(cfg config.AgentFlags, storage *storage.MemStorage, done <-chan struct{}) {
	tickerCollect := time.NewTicker(cfg.PollInterval)
	tickerSend := time.NewTicker(cfg.ReportInterval)
	defer tickerCollect.Stop()
	defer tickerSend.Stop()

	ctx := context.Background()

	for {
		select {
		case <-tickerSend.C:
			fmt.Println("Sending metrics...")
			// Старый POST запрос
			// err := sendMetrics(cfg.ServerAddr, storage)
			// Новый POST запрос JSON
			// err := sendMetricsJSON(cfg.ServerAddr, storage)
			// Новый POST запрос JSON с batching
			err := sendMetricsBatchJSON(cfg.ServerAddr, storage)

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
			storage.UpdateMetric(ctx, getMetricGauge("Alloc", float64(memStats.Alloc)))
			storage.UpdateMetric(ctx, getMetricGauge("BuckHashSys", float64(memStats.BuckHashSys)))
			storage.UpdateMetric(ctx, getMetricGauge("Frees", float64(memStats.Frees)))
			storage.UpdateMetric(ctx, getMetricGauge("GCCPUFraction", memStats.GCCPUFraction))
			storage.UpdateMetric(ctx, getMetricGauge("GCSys", float64(memStats.GCSys)))
			storage.UpdateMetric(ctx, getMetricGauge("HeapAlloc", float64(memStats.HeapAlloc)))
			storage.UpdateMetric(ctx, getMetricGauge("HeapIdle", float64(memStats.HeapIdle)))
			storage.UpdateMetric(ctx, getMetricGauge("HeapInuse", float64(memStats.HeapInuse)))
			storage.UpdateMetric(ctx, getMetricGauge("HeapObjects", float64(memStats.HeapObjects)))
			storage.UpdateMetric(ctx, getMetricGauge("HeapReleased", float64(memStats.HeapReleased)))
			storage.UpdateMetric(ctx, getMetricGauge("HeapSys", float64(memStats.HeapSys)))
			storage.UpdateMetric(ctx, getMetricGauge("LastGC", float64(memStats.LastGC)/1e9))
			storage.UpdateMetric(ctx, getMetricGauge("Lookups", float64(memStats.Lookups)))
			storage.UpdateMetric(ctx, getMetricGauge("MCacheInuse", float64(memStats.MCacheInuse)))
			storage.UpdateMetric(ctx, getMetricGauge("MCacheSys", float64(memStats.MCacheSys)))
			storage.UpdateMetric(ctx, getMetricGauge("MSpanInuse", float64(memStats.MSpanInuse)))
			storage.UpdateMetric(ctx, getMetricGauge("MSpanSys", float64(memStats.MSpanSys)))
			storage.UpdateMetric(ctx, getMetricGauge("Mallocs", float64(memStats.Mallocs)))
			storage.UpdateMetric(ctx, getMetricGauge("NextGC", float64(memStats.NextGC)))
			storage.UpdateMetric(ctx, getMetricGauge("NumForcedGC", float64(memStats.NumForcedGC)))
			storage.UpdateMetric(ctx, getMetricGauge("NumGC", float64(memStats.NumGC)))
			storage.UpdateMetric(ctx, getMetricGauge("OtherSys", float64(memStats.OtherSys)))
			storage.UpdateMetric(ctx, getMetricGauge("PauseTotalNs", float64(memStats.PauseTotalNs)/1e9))
			storage.UpdateMetric(ctx, getMetricGauge("StackInuse", float64(memStats.StackInuse)))
			storage.UpdateMetric(ctx, getMetricGauge("StackSys", float64(memStats.StackSys)))
			storage.UpdateMetric(ctx, getMetricGauge("Sys", float64(memStats.Sys)))
			storage.UpdateMetric(ctx, getMetricGauge("TotalAlloc", float64(memStats.TotalAlloc)))
			// Дополнительные
			storage.UpdateMetric(ctx, getMetricGauge("RandomValue", float64(rand.IntN(100))))
			pollCnt := storage.GetMetric(ctx, "PollCount")
			if pollCnt == nil {
				storage.UpdateMetric(ctx, getMetricCount("PollCount", 1))
			} else {
				storage.UpdateMetric(ctx, getMetricCount("PollCount", *pollCnt.Delta+1))
			}
			fmt.Println("Collecting metrics done")
		case <-done:
			fmt.Println("Stop processing metrics")
			return
		}
	}
}
