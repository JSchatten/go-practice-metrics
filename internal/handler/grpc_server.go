package handler

import (
	"context"
	"fmt"

	"github.com/JSchatten/go-practice-metrics/genproto/proto"
	"github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer реализует интерфейс MetricsServer из сгенерированного protobuf.
type GRPCServer struct {
	proto.UnimplementedMetricsServer
	storage storage.Storage
}

// NewGRPCServer создает новый gRPC-сервер для метрик.
func NewGRPCServer(storage storage.Storage) *GRPCServer {
	return &GRPCServer{
		storage: storage,
	}
}

// UpdateMetrics обрабатывает запрос на обновление метрик.
//
// Реализует метод интерфейса MetricsServer.
// Принимает запрос с батчем метрик, проверяет IP-адрес из метаданных
// и сохраняет метрики в хранилище.
//
// Возвращает:
//   - *proto.UpdateMetricsResponse: пустой ответ при успехе
//   - error: в случае ошибки (например, PermissionDenied при проверке IP)
func (s *GRPCServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	// Преобразуем метрики из proto в внутреннюю модель
	metrics := make([]model.Metrics, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		metric, err := protoToModelMetric(m)
		if err != nil {
			log.Error().Err(err).Msg("failed to convert metric")
			// Пропускаем некорректные метрики
			continue
		}
		metrics = append(metrics, *metric)
	}

	// Сохраняем метрики в хранилище
	if err := s.storage.UpdateMetricBatch(ctx, &metrics); err != nil {
		log.Error().Err(err).Msg("failed to update metrics in storage")
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	log.Info().Int("count", len(metrics)).Msg("metrics updated via gRPC")
	return &proto.UpdateMetricsResponse{}, nil
}

// protoToModelMetric конвертирует метрику из protobuf-формата в внутреннюю модель.
func protoToModelMetric(p *proto.Metric) (*model.Metrics, error) {
	m := &model.Metrics{ID: p.Id}
	switch p.Type {
	case proto.Metric_GAUGE:
		m.MType = model.Gauge
		m.Value = &p.Value
	case proto.Metric_COUNTER:
		m.MType = model.Counter
		m.Delta = &p.Delta
	default:
		return nil, fmt.Errorf("unknown metric type: %v", p.Type)
	}
	return m, nil
}
