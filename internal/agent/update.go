package agent

import (
	"context"
	"log"
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

func addError(errorsUpdating *[]error, err error) {
	if err != nil {
		*errorsUpdating = append(*errorsUpdating, err)
	}
}

func UpdateRuntimeMetrics(cfg config.AgentFlags, storage *storage.MemStorage, done <-chan struct{}) {
	tickerCollect := time.NewTicker(cfg.PollInterval)
	tickerSend := time.NewTicker(cfg.ReportInterval)
	defer tickerCollect.Stop()
	defer tickerSend.Stop()

	ctx := context.Background()

	httpClient := CreateHTTPClient()

	for {
		select {
		case <-tickerSend.C:
			log.Println("Sending metrics...")
			// Старый POST запрос
			// err := sendMetrics(httpClient, cfg.ServerAddr, storage)
			// Старый POST запрос JSON
			// err := sendMetricsJSON(httpClient, cfg.ServerAddr, storage)
			// Новый POST запрос JSON с batching, gzip, hash
			err := sendMetricsBatchJSON(httpClient, cfg.ServerAddr, storage, cfg.HashKey)

			if err != nil {
				log.Printf("Error sending metrics: %v\n", err)
			} else {
				log.Println("Sended successful")
			}
		case <-tickerCollect.C:
			log.Println("Collecting metrics...")
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			// Большой список, 1e9 для перевода в секунды
			var errorsUpdating []error
			// Выглядит несуразно, но работает; думаю в будущем выделить в отдльеный объект
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("Alloc", float64(memStats.Alloc))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("Alloc", float64(memStats.Alloc))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("BuckHashSys", float64(memStats.BuckHashSys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("Frees", float64(memStats.Frees))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("GCCPUFraction", memStats.GCCPUFraction)))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("GCSys", float64(memStats.GCSys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("HeapAlloc", float64(memStats.HeapAlloc))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("HeapIdle", float64(memStats.HeapIdle))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("HeapInuse", float64(memStats.HeapInuse))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("HeapObjects", float64(memStats.HeapObjects))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("HeapReleased", float64(memStats.HeapReleased))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("HeapSys", float64(memStats.HeapSys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("LastGC", float64(memStats.LastGC)/1e9)))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("Lookups", float64(memStats.Lookups))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("MCacheInuse", float64(memStats.MCacheInuse))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("MCacheSys", float64(memStats.MCacheSys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("MSpanInuse", float64(memStats.MSpanInuse))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("MSpanSys", float64(memStats.MSpanSys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("Mallocs", float64(memStats.Mallocs))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("NextGC", float64(memStats.NextGC))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("NumForcedGC", float64(memStats.NumForcedGC))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("NumGC", float64(memStats.NumGC))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("OtherSys", float64(memStats.OtherSys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("PauseTotalNs", float64(memStats.PauseTotalNs)/1e9)))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("StackInuse", float64(memStats.StackInuse))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("StackSys", float64(memStats.StackSys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("Sys", float64(memStats.Sys))))
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("TotalAlloc", float64(memStats.TotalAlloc))))

			// Дополнительные
			addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricGauge("RandomValue", float64(rand.IntN(100)))))

			pollCnt := storage.GetMetric(ctx, "PollCount")
			if pollCnt == nil {
				addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricCount("PollCount", 1)))
			} else {
				addError(&errorsUpdating, storage.UpdateMetric(ctx, getMetricCount("PollCount", *pollCnt.Delta+1)))
			}
			if len(errorsUpdating) > 0 {
				for _, err := range errorsUpdating {
					log.Printf("Error updating metric: %v", err)
				}
				errorsUpdating = []error{}
			}
			log.Println("Collecting metrics done")
		case <-done:
			log.Println("Stop processing metrics")
			return
		}
	}
}
