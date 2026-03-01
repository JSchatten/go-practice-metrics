package agent

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/JSchatten/go-practice-metrics/internal/crypto"
	"github.com/JSchatten/go-practice-metrics/internal/gzip"
	hashprocess "github.com/JSchatten/go-practice-metrics/internal/hashprocess"
	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

type workerPool struct {
	serverAddr string
	hashKey    string
	cryptoKey  string
	limiter    *rate.Limiter
	batchCh    chan []MetricsModel.Metrics
	ctx        context.Context
	wg         sync.WaitGroup
	client     *resty.Client
}

func newWorkerPool(
	client *resty.Client,
	serverAddr, hashKey, cryptoKey string,
	limiter *rate.Limiter,
	batchCh chan []MetricsModel.Metrics,
	ctx context.Context,

) *workerPool {
	return &workerPool{
		client:     client,
		serverAddr: serverAddr,
		hashKey:    hashKey,
		cryptoKey:  cryptoKey,
		limiter:    limiter,
		batchCh:    batchCh,
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
	close(wp.batchCh)
	wp.wg.Wait()
}

func (wp *workerPool) worker() {
	defer wp.wg.Done()

	for batch := range wp.batchCh { // range по batchCh
		wp.sendBatch(batch)
	}

}

func (wp *workerPool) SendBatch(batch []MetricsModel.Metrics) {
	select {
	case wp.batchCh <- batch:
	case <-wp.ctx.Done():
	}
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

	// прямое шифрование падает при больших батчах, нужно добавить AES
	// ERR Failed to encrypt request body error="crypto/rsa: message too long for RSA key
	log.Info().Int("json_size", len(jsonData)).Msg("JSON size before encryption")
	// Шифруем тело запроса, если указан путь к публичному ключу
	var bodyData []byte = jsonData
	if wp.cryptoKey != "" {
		log.Info().Msgf("Using public key for encryption: %s", wp.cryptoKey)
		pubKey, err := crypto.LoadRSAPublicKey(wp.cryptoKey)
		if err != nil {
			log.Error().Err(err).Msg("Failed to load public key")
			return
		}

		encryptedData, err := crypto.HybridEncrypt(pubKey, jsonData)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hybrid encrypt request body")
			return
		}
		bodyData = []byte(encryptedData)
		log.Info().Msgf("Successfully hybrid encrypted %d bytes of data", len(encryptedData))
	}

	compressed, err := gzip.CompressGZIP(bodyData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to compress batch")
		return
	}

	request := wp.client.R().SetContext(wp.ctx)

	if wp.hashKey != "" {
		sign := hashprocess.Sign(compressed, wp.hashKey)
		request.SetHeader("HashSHA256", sign)
	}

	// Устанавливаем заголовки и тело после обработки
	request = request.SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressed)

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
