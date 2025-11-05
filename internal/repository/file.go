package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
)

type FileRepository struct {
	filePath     string
	mxFileAccess sync.Mutex // защищаем доступ к файлу
}

func NewFileRepository(filePath string) *FileRepository {
	return &FileRepository{
		filePath: filePath,
	}
}

func (r *FileRepository) SaveMetrics(metrics map[string]*model.Metrics) error {
	if r.filePath == "" {
		return nil
	}

	r.mxFileAccess.Lock()
	defer r.mxFileAccess.Unlock()

	// Делаем снимок
	snapshot := make([]model.Metrics, 0, len(metrics))
	for _, m := range metrics {
		snapshot = append(snapshot, *m)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return os.WriteFile(r.filePath, data, 0600)
}

func (r *FileRepository) LoadMetrics() (map[string]*model.Metrics, error) {
	if r.filePath == "" {
		return make(map[string]*model.Metrics), nil
	}

	r.mxFileAccess.Lock()
	defer r.mxFileAccess.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*model.Metrics), nil
		}
		return nil, fmt.Errorf("failed to read metrics file: %w", err)
	}

	var metrics []model.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	result := make(map[string]*model.Metrics)
	for i := range metrics {
		m := metrics[i]
		result[m.ID] = &m
	}

	return result, nil
}
