package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	models "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_LoadMetrics_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	repo := NewFileRepository(filepath.Join(dir, "nonexistent.json"))

	data, err := repo.LoadMetrics()
	require.NoError(t, err)
	assert.Equal(t, []byte{}, data)
}

func TestFileRepository_LoadMetrics_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.json")
	err := os.WriteFile(filePath, []byte{}, 0600)
	require.NoError(t, err)

	repo := NewFileRepository(filePath)
	data, err := repo.LoadMetrics()
	require.NoError(t, err)
	assert.Equal(t, []byte{}, data)
}

func TestFileRepository_LoadMetrics_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "metrics.json")

	// Подготовим данные
	metrics := []models.Metrics{
		{
			ID:    "cpu",
			MType: "gauge",
			Value: floatPtr(99.5),
		},
		{
			ID:    "requests",
			MType: "counter",
			Delta: intPtr(100),
		},
	}
	data, err := json.MarshalIndent(metrics, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(filePath, data, 0600)
	require.NoError(t, err)

	repo := NewFileRepository(filePath)
	loaded, err := repo.LoadMetrics()
	require.NoError(t, err)
	assert.JSONEq(t, string(data), string(loaded))
}

func TestFileRepository_LoadMetrics_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "broken.json")
	err := os.WriteFile(filePath, []byte(`{ "invalid": `), 0600)
	require.NoError(t, err)

	repo := NewFileRepository(filePath)
	_, err = repo.LoadMetrics()
	require.NoError(t, err)
}

func TestFileRepository_SaveMetrics_Success(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "save.json")

	repo := NewFileRepository(filePath)
	metrics := []models.Metrics{
		{
			ID:    "test_gauge",
			MType: "gauge",
			Value: floatPtr(42.1),
		},
	}
	data, err := json.MarshalIndent(metrics, "", "  ")
	require.NoError(t, err)

	err = repo.SaveMetrics(data)
	require.NoError(t, err)

	// Проверим, что файл создался и содержит данные
	loaded, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.JSONEq(t, string(data), string(loaded))
}

func TestFileRepository_SaveMetrics_EmptyData(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.json")

	repo := NewFileRepository(filePath)
	err := repo.SaveMetrics([]byte{})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNoData)
}

func TestFileRepository_SaveMetrics_NoFilePath(t *testing.T) {
	repo := NewFileRepository("") // пустой путь
	err := repo.SaveMetrics([]byte(`{"test":1}`))
	require.NoError(t, err) // должна быть "тихая" операция
}

func TestFileRepository_FilePath_Method(t *testing.T) {
	repo := NewFileRepository("/tmp/test.json")
	assert.Equal(t, "/tmp/test.json", repo.FilePath())

	var nilRepo *FileRepository
	assert.Equal(t, "", nilRepo.FilePath()) // защита от nil
}

// Вспомогательные функции
func intPtr(i int64) *int64 {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
