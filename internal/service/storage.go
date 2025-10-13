package service

import (
	"encoding/json"
	"fmt"
	"sync"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
)

type MemStorage struct {
	Metrics map[string]*model.Metrics
	mx      sync.RWMutex // Мютекс Rw, для параллельнго чтения
}

type Storage interface {
	UpdateMetric(metric *model.Metrics) error
	GetMetric(id string) *model.Metrics
	String() string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*model.Metrics),
	}
}

func (s *MemStorage) String() string {
	s.mx.RLock()
	defer s.mx.RUnlock()

	if len(s.Metrics) == 0 {
		return ""
	} else {
		var metrics = []model.Metrics{}

		for _, metric := range s.Metrics {
			metrics = append(metrics, *metric)
		}

		metricsStr, err := json.Marshal(metrics)
		if err != nil {
			panic(err)
		} else {
			return string(metricsStr)
		}
	}
}

func (s *MemStorage) UpdateMetric(metric *model.Metrics) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("value is required for gauge")
		}
		s.Metrics[metric.ID] = &model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
	case model.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("delta is required for counter")
		}
		existing, exists := s.Metrics[metric.ID]
		if exists && existing.MType == model.Counter {
			// Добавляем новое значение к существующему
			newDelta := *existing.Delta + *metric.Delta
			s.Metrics[metric.ID] = &model.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: &newDelta,
			}
		} else {
			// Создаём новую метрику
			s.Metrics[metric.ID] = &model.Metrics{
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

func (s *MemStorage) GetMetric(id string) *model.Metrics {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if metric, exists := s.Metrics[id]; exists {
		return metric
	} else {
		return nil
	}
}
