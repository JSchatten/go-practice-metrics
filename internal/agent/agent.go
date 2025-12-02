package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/JSchatten/go-practice-metrics/internal/config"
	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/time/rate"
)

const cnstBatchSize = 20

type Agent struct {
	config        *config.AgentFlags
	storage       *service.MemStorage
	httpClient    *resty.Client
	tickerCollect *time.Ticker
	done          chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	rateLimiter *rate.Limiter
}

func NewAgent(cfg *config.AgentFlags) (*Agent, error) {
	storage, err := service.NewMemStorage("", 0, false, "")
	if err != nil {
		return nil, err
	}

	client := resty.New()
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		config:        cfg,
		storage:       storage,
		httpClient:    client,
		tickerCollect: time.NewTicker(cfg.PollInterval),
		done:          make(chan struct{}),
		ctx:           ctx,
		cancel:        cancel,

		rateLimiter: rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateLimit),
	}, nil
}

func (a *Agent) Start() {
	log.Println("Agent is running...")

	batchCh := make(chan []MetricsModel.Metrics, 100)

	workerPool := newWorkerPool(
		a.httpClient,
		a.config.ServerAddr,
		a.config.HashKey,
		a.rateLimiter,
		batchCh,
		a.ctx,
	)
	workerPool.start(a.config.RateLimit) // число воркеров = rate limit

	// Откладываем завершение
	defer workerPool.wait()

	tickerReport := time.NewTicker(a.config.ReportInterval)
	defer tickerReport.Stop()

	for {
		select {
		case <-a.tickerCollect.C:
			a.collectMetrics(a.ctx)
		case <-tickerReport.C:
			a.sendAllMetrics(workerPool)
		case <-a.done:
			log.Println("Agent stopped.")
			return
		}
	}
}

func (a *Agent) Stop() {
	a.tickerCollect.Stop()
	close(a.done)
}

func (a *Agent) collectMetrics(ctx context.Context) {
	log.Println("Collecting metrics...")
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	// Большой список, 1e9 для перевода в секунды
	var errorsUpdating []error
	// Выглядит несуразно, но работает; думаю в будущем выделить в отдльеный объект
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("Alloc", float64(memStats.Alloc))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("Alloc", float64(memStats.Alloc))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("BuckHashSys", float64(memStats.BuckHashSys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("Frees", float64(memStats.Frees))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("GCCPUFraction", memStats.GCCPUFraction)))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("GCSys", float64(memStats.GCSys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("HeapAlloc", float64(memStats.HeapAlloc))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("HeapIdle", float64(memStats.HeapIdle))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("HeapInuse", float64(memStats.HeapInuse))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("HeapObjects", float64(memStats.HeapObjects))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("HeapReleased", float64(memStats.HeapReleased))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("HeapSys", float64(memStats.HeapSys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("LastGC", float64(memStats.LastGC)/1e9)))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("Lookups", float64(memStats.Lookups))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("MCacheInuse", float64(memStats.MCacheInuse))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("MCacheSys", float64(memStats.MCacheSys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("MSpanInuse", float64(memStats.MSpanInuse))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("MSpanSys", float64(memStats.MSpanSys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("Mallocs", float64(memStats.Mallocs))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("NextGC", float64(memStats.NextGC))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("NumForcedGC", float64(memStats.NumForcedGC))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("NumGC", float64(memStats.NumGC))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("OtherSys", float64(memStats.OtherSys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("PauseTotalNs", float64(memStats.PauseTotalNs)/1e9)))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("StackInuse", float64(memStats.StackInuse))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("StackSys", float64(memStats.StackSys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("Sys", float64(memStats.Sys))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("TotalAlloc", float64(memStats.TotalAlloc))))

	// Новые метрики для памяти
	v, _ := mem.VirtualMemory()
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("TotalMemory", float64(v.Total))))
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("FreeMemory", float64(v.Free))))
	// Новые метрики для CPU

	cpuUsesage, err := cpu.Percent(time.Duration(0), true)
	if err != nil {
		addError(&errorsUpdating, err)
		log.Printf("Error getting CPU usage: %v", err)
	} else {
		for cpuIndx, cpuUse := range cpuUsesage {
			addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge(fmt.Sprintf("CPUutilization%d", cpuIndx), cpuUse)))
		}
	}

	// Дополнительные
	addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricGauge("RandomValue", float64(rand.IntN(100)))))

	pollCnt := a.storage.GetMetric(ctx, "PollCount")
	if pollCnt == nil {
		addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricCount("PollCount", 1)))
	} else {
		addError(&errorsUpdating, a.storage.UpdateMetric(ctx, getMetricCount("PollCount", *pollCnt.Delta+1)))
	}
	if len(errorsUpdating) > 0 {
		for _, err := range errorsUpdating {
			log.Printf("Error updating metric: %v", err)
		}
		errorsUpdating = []error{}
	}
	log.Println("Collecting metrics done")

}

func (a *Agent) sendAllMetrics(wp *workerPool) {
	ctx := context.Background()

	// Получаем все метрики из storage
	metrics := a.storage.GetAllMetrics(ctx)
	if len(metrics) == 0 {
		log.Println("No metrics to send")
		return
	}

	log.Printf("Sending %d metrics", len(metrics))

	// Делим на батчи и отправляем
	for i := 0; i < len(metrics); i += cnstBatchSize {
		end := i + cnstBatchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		batch := metrics[i:end]

		// Асинхронная отправка
		wp.SendBatch(batch)
	}
}
