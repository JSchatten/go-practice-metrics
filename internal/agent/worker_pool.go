package agent

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/JSchatten/go-practice-metrics/internal/gzip"
	hashprocess "github.com/JSchatten/go-practice-metrics/internal/hashprocess"
	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

const batchSize = 20

type workerPool struct {
	client     *resty.Client
	serverAddr string
	hashKey    string
	limiter    *rate.Limiter
	metricsCh  <-chan MetricsModel.Metrics
	ctx        context.Context
	wg         sync.WaitGroup
}

func newWorkerPool(
	client *resty.Client,
	serverAddr, hashKey string,
	limiter *rate.Limiter,
	metricsCh <-chan MetricsModel.Metrics,
	ctx context.Context,
) *workerPool {
	return &workerPool{
		client:     client,
		serverAddr: serverAddr,
		hashKey:    hashKey,
		limiter:    limiter,
		metricsCh:  metricsCh,
		ctx:        ctx,
	}
}

func (wp *workerPool) start(numWorkers int) {
	wp.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go wp.worker()
	}
}

func (wp *workerPool) wait() {
	wp.wg.Wait()
}

func (wp *workerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case metric, ok := <-wp.metricsCh:
			if !ok {
				return // канал закрыт
			}
			batch := wp.collectBatch(metric)
			wp.sendBatch(batch)
		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *workerPool) collectBatch(first MetricsModel.Metrics) []MetricsModel.Metrics {
	batch := []MetricsModel.Metrics{first}

	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	for len(batch) < batchSize {
		select {
		case metric, ok := <-wp.metricsCh:
			if !ok {
				return batch
			}
			batch = append(batch, metric)
		case <-timer.C:
			return batch
		case <-wp.ctx.Done():
			return batch
		}
	}
	return batch
}

// переписанная отпарвка по сути (что было в httpSendBatchJSON)
// и чтобы был прямо таки явный батчинг каналами, необходимо будет
// отказатья от схемы с коллектом каждую секунду и отправку на пятую итерацию
func (wp *workerPool) sendBatch(metrics []MetricsModel.Metrics) {
	if len(metrics) == 0 {
		return
	}

	// Ждём разрешения от rate limiter
	if err := wp.limiter.Wait(wp.ctx); err != nil {
		log.Warn().Err(err).Msg("Rate limiter wait cancelled")
		return
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal batch")
		return
	}

	compressed, err := gzip.CompressGZIP(jsonData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to compress batch")
		return
	}

	request := wp.client.R().
		SetContext(wp.ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressed)

	if wp.hashKey != "" {
		sign := hashprocess.Sign(compressed, wp.hashKey)
		request.SetHeader("HashSHA256", sign)
	}

	resp, err := request.Post("http://" + wp.serverAddr + "/updates/")
	if err != nil {
		log.Error().Err(err).Msg("Send request failed")
		return
	}
	if resp.StatusCode() != 200 {
		log.Warn().Int("status", resp.StatusCode()).Msg("Send rejected by server")
		return
	}

	log.Info().Int("count", len(metrics)).Msg("Batch sent successfully")
}
