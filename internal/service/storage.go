package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	filerepo "github.com/JSchatten/go-practice-metrics/internal/repository"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/rs/zerolog/log"
)

type MemStorage struct {
	Metrics      map[string]*model.Metrics
	mxDataAccess sync.RWMutex // Мютекс Rw, для параллельнго чтения
	fileRepo     *filerepo.FileRepository
}

type Storage interface {
	UpdateMetric(metric *model.Metrics) error
	GetMetric(id string) *model.Metrics
	String() string
}

func NewMemStorage(filePath string, flushInterval time.Duration, loadFromFile bool) *MemStorage {

	var fileRepo *filerepo.FileRepository
	if filePath != "" {
		fileRepo = filerepo.NewFileRepository(filePath)
	}

	storage := &MemStorage{
		Metrics:  make(map[string]*model.Metrics),
		fileRepo: fileRepo,
	}

	if loadFromFile {
		storage.loadFromDisk()
	}

	if fileRepo != nil {
		if loadFromFile {
			storage.loadFromDisk()
		}
		go storage.startAutoSave(filePath, flushInterval)
	} else {
		log.Logger.Warn().Msgf("Warning: no filepath for load data\n")
	}
	return storage
}

func (s *MemStorage) loadFromDisk() {
	metrics, err := s.fileRepo.LoadMetrics()
	if err != nil {
		log.Logger.Warn().Msgf("Warning: failed to load metrics from file: %v\n", err)
		return
	}

	s.mxDataAccess.Lock()
	defer s.mxDataAccess.Unlock()
	s.Metrics = metrics
}

func (s *MemStorage) startAutoSave(filepath string, interval time.Duration) {
	if interval <= 0 {
		log.Logger.Info().Msg("Ignore interval")
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Logger.Info().Msg("Starting auto-save")

	for range ticker.C {
		if err := s.SaveToFile(filepath); err != nil {
			log.Logger.Error().Err(err).Msgf("Failed to save metrics to file: %v\n", err)
		}
	}
}

func (s *MemStorage) SaveToFile(filepath string) error {
	s.mxDataAccess.RLock()
	defer s.mxDataAccess.RUnlock()

	if s.fileRepo != nil {
		if err := s.fileRepo.SaveMetrics(s.Metrics); err != nil {
			fmt.Printf("Failed to save metrics to file: %v\n", err)
		}
	} else {
		log.Logger.Info().Msgf("Ignore save to file, because file path is empty")
	}

	return nil
}

func (s *MemStorage) String() string {
	s.mxDataAccess.RLock()
	defer s.mxDataAccess.RUnlock()

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
	s.mxDataAccess.Lock()
	defer s.mxDataAccess.Unlock()

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
	s.mxDataAccess.RLock()
	defer s.mxDataAccess.RUnlock()
	if metric, exists := s.Metrics[id]; exists {
		return metric
	} else {
		return nil
	}
}
