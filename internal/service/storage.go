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
	UpdateMetricBatch(ctx context.Context, metric *[]model.Metrics) error
	GetMetric(ctx context.Context, id string) *model.Metrics
	String() string
	PingDatabase(ctx context.Context) error
}

func NewMemStorage(filePath string, flushInterval time.Duration, loadFromFile bool, postgresDSN string) (*MemStorage, error) {

	var fileRepo *fileRepos.FileRepository
	if filePath != "" {
		fileRepo = fileRepos.NewFileRepository(filePath)
	}

	var repoDBConnected = false

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
			repoDBConnected = true
		}
	}

	storage := &MemStorage{
		Metrics:  make(map[string]*model.Metrics),
		fileRepo: fileRepo,
		dbRepo:   repo,
	}

	if fileRepo != nil && !repoDBConnected {
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

	err := s.updateMetricInMemory(metric)
	if err != nil {
		log.Logger.Error().Err(err).Msg(ErrMetricSaveMemory.Error())
		return err
	}

	if s.PingDatabase(ctx) == nil {
		if err := s.dbRepo.UpdateMetric(ctx, s.Metrics[metric.ID]); err != nil {
			log.Logger.Error().Err(err).Msg(ErrMetricSaveFailedDatabase.Error())
		}
	}

	if s.immediatelyFlush {
		if err := s.SaveToFile(); err != nil {
			log.Logger.Error().Err(err).Msg(ErrMetricSaveFailedFile.Error())
		}
	}
	return nil
}

func (s *MemStorage) UpdateMetricBatch(ctx context.Context, metrics *[]model.Metrics) error {
	s.mxDataAccess.Lock()
	defer s.mxDataAccess.Unlock()

	for _, elem := range *metrics {
		err := s.updateMetricInMemory(&elem)
		if err != nil {
			log.Logger.Error().Err(err).Msg(ErrMetricSaveMemory.Error())
			return err
		}
	}

	if s.PingDatabase(ctx) == nil {
		if err := s.dbRepo.UpdateMetricBatch(ctx, *metrics); err != nil {
			log.Logger.Error().Err(err).Msg(ErrMetricSaveFailedDatabase.Error())
		}
	}

	if s.immediatelyFlush {
		if err := s.SaveToFile(); err != nil {
			log.Logger.Error().Err(err).Msg(ErrMetricSaveFailedFile.Error())
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
			// TODO убрать после проверки
			// s.Metrics[metric.ID] = metric
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

func (s *MemStorage) updateMetricInMemory(metricIn *model.Metrics) error {
	switch metricIn.MType {
	case model.Gauge:
		if metricIn.Value == nil {
			return ErrValueRequired
		}
		updated := &model.Metrics{
			ID:    metricIn.ID,
			MType: metricIn.MType,
			Value: metricIn.Value,
		}
		s.Metrics[metricIn.ID] = updated
		return nil

	case model.Counter:
		if metricIn.Delta == nil {
			return ErrDeltaRequired
		}
		existing, exists := s.Metrics[metricIn.ID]
		if exists && existing.MType == model.Counter {
			newDelta := *existing.Delta + *metricIn.Delta
			updated := &model.Metrics{
				ID:    metricIn.ID,
				MType: metricIn.MType,
				Delta: &newDelta,
			}
			s.Metrics[metricIn.ID] = updated
			return nil
		} else {
			updated := &model.Metrics{
				ID:    metricIn.ID,
				MType: metricIn.MType,
				Delta: metricIn.Delta,
			}
			s.Metrics[metricIn.ID] = updated
			return nil
		}

	default:
		return NewErrUnknownMetricType(string(metricIn.MType))
	}
}
