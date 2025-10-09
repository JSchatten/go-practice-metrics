package service

import (
	"fmt"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
)

type MemStorage struct {
	metrics map[string]*model.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]*model.Metrics),
	}
}

func (s *MemStorage) UpdateMetric(metric *model.Metrics) error {
	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("value is required for gauge")
		}
		s.metrics[metric.ID] = &model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
	case model.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("delta is required for counter")
		}
		existing, exists := s.metrics[metric.ID]
		if exists && existing.MType == model.Counter {
			// Добавляем новое значение к существующему
			newDelta := *existing.Delta + *metric.Delta
			s.metrics[metric.ID] = &model.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: &newDelta,
			}
		} else {
			// Создаём новую метрику
			s.metrics[metric.ID] = &model.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: metric.Delta,
			}
		}
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}
	return nil
}
