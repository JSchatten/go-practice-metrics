// Package service реализует бизнес-логику управления метриками.
//
// Основные функции:
//   - Хранение метрик в памяти
//   - Поддержка сохранения в файл (с периодической автосохранением)
//   - Интеграция с PostgreSQL (опционально)
//   - Приоритет: БД > память > файл
//   - Потокобезопасный доступ через sync.RWMutex
//
// Поведение при запуске:
//   - Если задан DSN и БД доступна - все операции идут через неё
//   - Иначе используется локальное хранилище (память + файл)
//   - При наличии файла и флага -restore - метрики загружаются при старте
//
// Пакет реализует паттерн "Repository", делегируя низкоуровневые операции:
//   - fileRepos.FileRepository - работа с файлом
//   - postgresqlRepo.MetricRepo - работа с PostgreSQL
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

// MemStorage - реализация хранилища метрик в памяти с поддержкой:
//   - Потокобезопасного доступа (через RWMutex)
//   - Сохранения на диск
//   - Работы с PostgreSQL
//
// Данные хранятся в виде map[string]*model.Metrics.
// Все методы, изменяющие состояние, блокируют запись; читающие - блокируют на чтение.
type MemStorage struct {
	// Metrics - основное хранилище метрик в памяти.
	Metrics          map[string]*model.Metrics
	mxDataAccess     sync.RWMutex // Мютекс Rw, для параллельнго чтения
	fileRepo         *fileRepos.FileRepository
	dbRepo           *postgresqlRepo.MetricRepo
	immediatelyFlush bool
}

// Storage - интерфейс, определяющий контракт для хранилища метрик.
// Позволяет использовать разные реализации (например, mock в тестах).
type Storage interface {
	// UpdateMetric обновляет или создаёт метрику.
	// Для counter - значение инкрементируется.
	// Для gauge - значение заменяется.
	// Поддерживает сохранение в БД и на диск.
	UpdateMetric(ctx context.Context, metric *model.Metrics) error

	// UpdateMetricBatch выполняет пакетное обновление метрик.
	// Обрабатывает все метрики в одной блокировке.
	// Сохраняет в БД и/или на диск после успешного обновления.
	UpdateMetricBatch(ctx context.Context, metric *[]model.Metrics) error

	// GetMetric возвращает метрику по имени.
	// Приоритет: БД > память.
	// Возвращает nil, если метрика не найдена.
	GetMetric(ctx context.Context, id string) *model.Metrics

	// String возвращает строковое представление всех метрик в формате JSON.
	// Используется, например, для логирования.
	// Паникует при ошибке сериализации (что маловероятно).
	String() string

	// PingDatabase проверяет доступность PostgreSQL.
	// Возвращает nil, если соединение активно.
	PingDatabase(ctx context.Context) error
}

// NewMemStorage создаёт новое хранилище с поддержкой:
//   - Локального файла (если filePath задан)
//   - PostgreSQL (если postgresDSN задан и доступен)
//
// Поведение:
//   - Если БД недоступна, используется файловое хранилище.
//   - Если loadFromFile == true и filePath задан - загружает метрики с диска.
//   - Если flushInterval > 0 - запускает горутину для периодического сохранения.
//   - Если flushInterval == 0 - включает immediate flush (сохранение после каждого обновления).
//
// Возвращает:
//   - Указатель на *MemStorage и nil в случае успеха
//   - nil и ошибку, если не удалось загрузить данные с диска
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

// PingDatabase проверяет соединение с PostgreSQL.
//
// Реализует интерфейс Storage.
//
// Возвращает:
//   - nil, если соединение с БД активно
//   - ошибку, если соединение разорвано или не было установлено
func (s *MemStorage) PingDatabase(ctx context.Context) error {
	return s.dbRepo.Ping(ctx)
}

// String возвращает JSON-представление всех метрик.
//
// Потокобезопасен (использует RLock).
//
// Поведение:
//   - Возвращает пустую строку, если метрик нет
//   - Сериализует все метрики в массив
//   - При ошибке сериализации - логирует и паникует (не должно происходить)
//
// Возвращает:
//   - JSON-строку с метриками
//   - Пустую строку, если метрик нет
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

// UpdateMetric обновляет одну метрику.
//
// Потокобезопасен (использует Lock).
//
// Логика:
//   - Обновляет в памяти (вызывает updateMetricInMemory)
//   - Если БД доступна - отправляет туда
//   - Если включён immediate flush - сохраняет в файл
//
// Возвращает:
//   - nil при успехе
//   - Ошибку валидации (например, ErrValueRequired)
func (s *MemStorage) UpdateMetric(ctx context.Context, metric *model.Metrics) error {
	s.mxDataAccess.Lock()
	defer s.mxDataAccess.Unlock()

	err := s.updateMetricInMemory(metric)
	if err != nil {
		log.Logger.Error().Err(err).Msg(ErrMetricSaveMemory.Error())
		return err
	}

	if s.PingDatabase(ctx) == nil {
		// if err := s.dbRepo.UpdateMetric(ctx, s.Metrics[metric.ID]); err != nil {
		if err := s.dbRepo.UpdateMetric(ctx, metric); err != nil {
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

// UpdateMetricBatch обновляет несколько метрик за одну транзакцию.
//
// Потокобезопасен (одна блокировка на весь батч).
//
// Логика:
//   - Последовательно обновляет каждую метрику в памяти
//   - При первой ошибке валидации - прерывает выполнение
//   - Если БД доступна - отправляет весь батч
//   - Если включён immediate flush - сохраняет всё в файл
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

// GetMetric возвращает метрику по имени.
//
// Потокобезопасен (использует RLock).
//
// Приоритет источника:
//  1. PostgreSQL (если доступна)
//  2. Память
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

// updateMetricInMemory обновляет метрику в карте Metrics.
//
// Не потокобезопасен - должен вызываться под Lock.
//
// Поведение:
//   - Gauge: заменяет значение
//   - Counter: прибавляет delta к существующему значению
func (s *MemStorage) updateMetricInMemory(metricIn *model.Metrics) error {
	// принимаем новое значение
	// расчет нового - дело уже на уровне репозитория
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

// GetAllMetrics возвращает копию всех метрик в виде среза.
//
// Потокобезопасен (использует RLock).
//
// Используется, например, для отображения на главной странице.
//
// Возвращает:
//   - Срез model.Metrics - всегда новый (не ссылка на внутренние данные)
func (s *MemStorage) GetAllMetrics(ctx context.Context) []model.Metrics {
	s.mxDataAccess.RLock()
	defer s.mxDataAccess.RUnlock()

	var metrics []model.Metrics
	for _, m := range s.Metrics {
		metrics = append(metrics, *m)
	}
	return metrics
}
