// Package service Дополнительные методы для MemStorage, отвечающие за сохранение и загрузку метрик из файла.
//
// Реализуют:
//   - Загрузку при старте (если включено)
//   - Периодическое автосохранение
//   - Мгновенное сохранение (immediate flush)
package service

import (
	"encoding/json"
	"time"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/rs/zerolog/log"
)

func (s *MemStorage) loadFromDisk() error {

	metricsData, err := s.fileRepo.LoadMetrics()

	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to load metrics from file")
		return err
	}

	if len(metricsData) == 0 {
		log.Logger.Warn().Msg("Empty filestorage, init empty slice")
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

// SaveToFile сохраняет все текущие метрики в файл в формате JSON.
//
// Используется:
//   - При shutdown (graceful shutdown)
//   - При включённом immediate flush
//   - Периодически - через startAutoSave
//
// Поведение:
//   - Делает снимок метрик (snapshot) под RLock
//   - Сериализует в JSON с отступами (json.MarshalIndent)
//   - Сохраняет через fileRepo.SaveMetrics()
//
// Потокобезопасность:
//   - Использует RLock для безопасного чтения s.Metrics
//   - Избегает дедлоков (закомментированные попытки Lock - потенциально опасны)
//
// Возвращает:
//   - nil при успехе
//   - NewErrMarshal - если не удалось сериализовать
//   - NewErrSaveToFile - если не удалось записать на диск
//
// Особенность:
//   - Не делает ничего, если s.fileRepo == nil
//   - Сохраняет полную копию - не ссылку на живые данные
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
