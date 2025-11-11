package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	fileRepos "github.com/JSchatten/go-practice-metrics/internal/repository/file"
	postgresqlRepo "github.com/JSchatten/go-practice-metrics/internal/repository/postgresql"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/rs/zerolog/log"
)

type MemStorage struct {
	Metrics          map[string]*model.Metrics
	mxDataAccess     sync.RWMutex // Мютекс Rw, для параллельнго чтения
	fileRepo         *fileRepos.FileRepository
	dbRepo           *postgresqlRepo.MetricRepo
	immediatelyFlush bool
}

type Storage interface {
	UpdateMetric(ctx context.Context, metric *model.Metrics) error
	GetMetric(ctx context.Context, id string) *model.Metrics
	String() string
	PingDatabase(ctx context.Context) error
}

func NewMemStorage(filePath string, flushInterval time.Duration, loadFromFile bool, postgresDSN string) (*MemStorage, error) {

	var fileRepo *fileRepos.FileRepository
	if filePath != "" {
		fileRepo = fileRepos.NewFileRepository(filePath)
	}

	repo, err := postgresqlRepo.NewMetricRepo(postgresDSN)
	if err != nil {
		log.Logger.Warn().Err(err).Msg("Failed to connect to postgres")
	} else {
		ctx := context.Background()
		err = repo.Ping(ctx)
		if err != nil {
			log.Logger.Warn().Err(err).Msg("No postgres connection, skip processing with database")
		} else {
			if err := repo.Migrate(ctx); err != nil {
				log.Logger.Warn().Err(err).Msg("Failed to migrate database")
			}
		}
	}

	storage := &MemStorage{
		Metrics:  make(map[string]*model.Metrics),
		fileRepo: fileRepo,
		dbRepo:   repo,
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

func (s *MemStorage) PingDatabase(ctx context.Context) error {
	return s.dbRepo.Ping(ctx)
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

func (s *MemStorage) UpdateMetric(ctx context.Context, metric *model.Metrics) error {
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
	if s.PingDatabase(ctx) == nil {
		if err := s.dbRepo.UpdateMetric(ctx, s.Metrics[metric.ID]); err != nil {
			log.Logger.Error().Err(err).Msgf("Failed to save metric to database")
		}
	}

	if s.immediatelyFlush {
		if err := s.SaveToFile(); err != nil {
			log.Logger.Error().Err(err).Msgf("Failed to save metrics to file")
		}
	}
	return nil
}

func (s *MemStorage) GetMetric(ctx context.Context, id string) *model.Metrics {
	s.mxDataAccess.RLock()
	defer s.mxDataAccess.RUnlock()

	// Возвращаем из БД, если она жива
	if s.PingDatabase(ctx) == nil {
		if metric, err := s.dbRepo.GetMetricByID(ctx, id); err == nil {
			s.Metrics[metric.ID] = metric
			return metric
		}
	}

	// Иначе берем из памят
	if metric, exists := s.Metrics[id]; exists {
		return metric
	} else {
		return nil
	}
}
