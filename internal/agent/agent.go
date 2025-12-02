package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/JSchatten/go-practice-metrics/internal/config"
	gzip "github.com/JSchatten/go-practice-metrics/internal/gzip"
	hashprocess "github.com/JSchatten/go-practice-metrics/internal/hashprocess"
	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/go-resty/resty/v2"
	logZero "github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type Agent struct {
	config        *config.AgentFlags
	storage       *service.MemStorage
	httpClient    *resty.Client
	tickerCollect *time.Ticker
	tickerSend    *time.Ticker
	done          chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func NewAgent(cfg *config.AgentFlags) (*Agent, error) {
	storage, err := service.NewMemStorage("", 0, false, "")
	if err != nil {
		return nil, err
	}

	client := resty.New() //.SetTimeout(3 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		config:        cfg,
		storage:       storage,
		httpClient:    client,
		tickerCollect: time.NewTicker(cfg.PollInterval),
		tickerSend:    time.NewTicker(cfg.ReportInterval),
		done:          make(chan struct{}),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

func (a *Agent) Start() {
	log.Println("Agent is running...")

	for {
		select {
		case <-a.tickerCollect.C:
			a.collectMetrics(a.ctx)

		case <-a.tickerSend.C:
			err := a.sendMetrics()
			if err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					logZero.Logger.Fatal().Err(err).Msg("Server closed, stopping agent")
					return
				} else {
					logZero.Logger.Fatal().Err(err).Msg("Error sending metrics")
				}
			}

		case <-a.done:
			log.Println("Agent stopped.")
			return
		}
	}
}

func (a *Agent) Stop() {
	a.tickerCollect.Stop()
	a.tickerSend.Stop()
	close(a.done)
}

func (a *Agent) sendMetrics() error {
	var sendingMetrics []MetricsModel.Metrics

	for _, metric := range a.storage.Metrics {
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

	request := a.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip")

	if a.config.HashKey != "" {
		signOfRequest := hashprocess.Sign(compressed, a.config.HashKey)
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

		resp, err := request.Post(fmt.Sprintf("http://%s/updates/", a.config.ServerAddr))

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
