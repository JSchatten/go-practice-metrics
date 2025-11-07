package service

import (
	"encoding/json"
	"sync"
	"time"

	filerepo "github.com/JSchatten/go-practice-metrics/internal/repository"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/rs/zerolog/log"
)

type MemStorage struct {
	Metrics          map[string]*model.Metrics
	mxDataAccess     sync.RWMutex // Мютекс Rw, для параллельнго чтения
	fileRepo         *filerepo.FileRepository
	immediatelyFlush bool
}

type Storage interface {
	UpdateMetric(metric *model.Metrics) error
	GetMetric(id string) *model.Metrics
	String() string
}

func NewMemStorage(filePath string, flushInterval time.Duration, loadFromFile bool) (*MemStorage, error) {

	var fileRepo *filerepo.FileRepository
	if filePath != "" {
		fileRepo = filerepo.NewFileRepository(filePath)
	}

	storage := &MemStorage{
		Metrics:  make(map[string]*model.Metrics),
		fileRepo: fileRepo,
	}

	if fileRepo != nil {
		if loadFromFile {
			err := storage.loadFromDisk()
			if err != nil {
				return nil, err
			}
		}
		if flushInterval > 0 {
			go storage.startAutoSave(flushInterval)
		} else {
			storage.immediatelyFlush = true
		}
	} else {
		log.Logger.Warn().Msgf("No filepath for load data")
	}
	return storage, nil
}

func (s *MemStorage) loadFromDisk() error {

	metricsData, err := s.fileRepo.LoadMetrics()

	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to load metrics from file")
		return err
	}

	if len(metricsData) == 0 {
		log.Logger.Info().Msg("Empty filestorage, init empty slice")
		metricsData = []byte("[]")
	}

	var metrics []model.Metrics
	if err := json.Unmarshal(metricsData, &metrics); err != nil {
		return NewErrLoadFromFile(s.fileRepo.FilePath(), err)
	}

	result := make(map[string]*model.Metrics)
	for i := range metrics {
		m := metrics[i]
		result[m.ID] = &m
	}

	s.mxDataAccess.Lock()
	defer s.mxDataAccess.Unlock()
	s.Metrics = result
	return nil
}

func (s *MemStorage) startAutoSave(interval time.Duration) {
	if interval <= 0 {
		log.Logger.Info().Msg("Ignore interval less than 0, gorutine was stopped")
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Logger.Info().Msg("Starting auto-save")

	for range ticker.C {
		if err := s.SaveToFile(); err != nil {
			log.Logger.Error().Err(err).Msgf("Failed to save metrics to file")
		}
	}
}

func (s *MemStorage) SaveToFile() error {
	// Потенциальный дедлок
	// if s.mxDataAccess.TryLock() {
	// 	defer s.mxDataAccess.RUnlock()
	// }
	// Потенциальный дедлок
	// s.mxDataAccess.RLock()
	// defer s.mxDataAccess.RUnlock()

	if s.fileRepo == nil {
		return nil
	}

	// Делаем снимок
	snapshot := make([]model.Metrics, 0, len(s.Metrics))
	for _, m := range s.Metrics {
		snapshot = append(snapshot, *m)
	}

	bytesToSave, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return NewErrMarshal(err)
	}

	if err := s.fileRepo.SaveMetrics(bytesToSave); err != nil {
		return NewErrSaveToFile(s.fileRepo.FilePath(), err)
	}

	return nil
}

func (s *MemStorage) String() string {
	s.mxDataAccess.RLock()
	defer s.mxDataAccess.RUnlock()

	if len(s.Metrics) == 0 {
		return ""
	}

	var metrics = []model.Metrics{}
	for _, metric := range s.Metrics {
		metrics = append(metrics, *metric)
	}
	metricsStr, err := json.Marshal(metrics)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to marshal metrics in String()")
		panic(NewErrMarshal(err))
	}

	return string(metricsStr)
}

func (s *MemStorage) UpdateMetric(metric *model.Metrics) error {
	s.mxDataAccess.Lock()
	defer s.mxDataAccess.Unlock()

	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return ErrValueRequired
		}
		s.Metrics[metric.ID] = &model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
	case model.Counter:
		if metric.Delta == nil {
			return ErrDeltaRequired
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
		return NewErrUnknownMetricType(string(metric.MType))
	}
	if s.immediatelyFlush {
		if err := s.SaveToFile(); err != nil {
			log.Logger.Error().Err(err).Msgf("Failed to save metrics to file")
		}
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
