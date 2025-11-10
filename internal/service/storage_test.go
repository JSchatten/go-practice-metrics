package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/stretchr/testify/suite"
)

// Вспомогательные функции
func intPtr(i int64) *int64 {
	v := i
	return &v
}

func floatPtr(f float64) *float64 {
	v := f
	return &v
}

// TestStorageSuite — набор тестов с общим setup/teardown
type TestStorageSuite struct {
	suite.Suite
	tempDir string
	file    string
}

func (s *TestStorageSuite) SetupTest() {
	// var err error
	s.tempDir = s.T().TempDir()
	s.file = filepath.Join(s.tempDir, "metrics.json")
}

func (s *TestStorageSuite) TearDownTest() {
	// Очистка выполняется автоматически через TempDir
}

func TestStorageSuiteRun(t *testing.T) {
	suite.Run(t, new(TestStorageSuite))
}

// --- Тесты ---

func (s *TestStorageSuite) TestNewMemStorage_NoFile() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)
	s.NotNil(storage)
	s.Nil(storage.fileRepo)
	s.NotNil(storage.Metrics)
	s.Equal(0, len(storage.Metrics))
}

func (s *TestStorageSuite) TestNewMemStorage_LoadFromFile_Exists() {
	// Подготовка файла
	metrics := []model.Metrics{
		{
			ID:    "cpu",
			MType: model.Gauge,
			Value: floatPtr(99.5),
		},
		{
			ID:    "requests",
			MType: model.Counter,
			Delta: intPtr(100),
		},
	}
	data, err := json.MarshalIndent(metrics, "", "  ")
	s.NoError(err)
	err = os.WriteFile(s.file, data, 0600)
	s.NoError(err)

	storage, err := NewMemStorage(s.file, 0, true, "")
	s.NoError(err)
	s.NotNil(storage.fileRepo)

	metric := storage.GetMetric("cpu")
	s.NotNil(metric)
	s.Equal("cpu", metric.ID)
	s.Equal(model.Gauge, metric.MType)
	s.InDelta(99.5, *metric.Value, 1e-6)

	metric = storage.GetMetric("requests")
	s.NotNil(metric)
	s.Equal(int64(100), *metric.Delta)
}

func (s *TestStorageSuite) TestNewMemStorage_LoadFromFile_NotExists() {
	storage, err := NewMemStorage(s.file, 0, true, "")
	s.NoError(err)
	s.NotNil(storage.fileRepo)
	s.Empty(storage.Metrics) // должно быть пусто
}

func (s *TestStorageSuite) TestNewMemStorage_LoadFromFile_InvalidJSON() {
	err := os.WriteFile(s.file, []byte(`{ "invalid": `), 0600)
	s.NoError(err)

	_, err = NewMemStorage(s.file, 0, true, "")
	s.Error(err)
}

func (s *TestStorageSuite) TestUpdateMetric_Gauge() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)

	metric := &model.Metrics{
		ID:    "cpu",
		MType: model.Gauge,
		Value: floatPtr(42.1),
	}
	err = storage.UpdateMetric(metric)
	s.NoError(err)

	got := storage.GetMetric("cpu")
	s.NotNil(got)
	s.Equal(model.Gauge, got.MType)
	s.InDelta(42.1, *got.Value, 1e-6)
}

func (s *TestStorageSuite) TestUpdateMetric_Counter_Increment() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)

	// Первое значение
	err = storage.UpdateMetric(&model.Metrics{
		ID:    "requests",
		MType: model.Counter,
		Delta: intPtr(5),
	})
	s.NoError(err)

	// Второе значение — должно прибавиться
	err = storage.UpdateMetric(&model.Metrics{
		ID:    "requests",
		MType: model.Counter,
		Delta: intPtr(3),
	})
	s.NoError(err)

	got := storage.GetMetric("requests")
	s.NotNil(got)
	s.Equal(int64(8), *got.Delta)
}

func (s *TestStorageSuite) TestUpdateMetric_UnknownType() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)

	err = storage.UpdateMetric(&model.Metrics{
		ID:    "bad",
		MType: "unknown",
		Value: floatPtr(1.0),
	})
	s.Error(err)
}

func (s *TestStorageSuite) TestUpdateMetric_NilValueOrDelta() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)

	err = storage.UpdateMetric(&model.Metrics{
		ID:    "cpu",
		MType: model.Gauge,
		Value: nil,
	})
	s.Error(err)
	s.Equal(ErrValueRequired, err)

	err = storage.UpdateMetric(&model.Metrics{
		ID:    "requests",
		MType: model.Counter,
		Delta: nil,
	})
	s.Error(err)
	s.Equal(ErrDeltaRequired, err)
}

func (s *TestStorageSuite) TestGetMetric_NotFound() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)

	got := storage.GetMetric("unknown")
	s.Nil(got)
}

func (s *TestStorageSuite) TestSaveToFile_ImmediatelyFlush() {
	storage, err := NewMemStorage(s.file, 0, false, "")
	s.NoError(err)
	storage.immediatelyFlush = true // принудительно

	err = storage.UpdateMetric(&model.Metrics{
		ID:    "cpu",
		MType: model.Gauge,
		Value: floatPtr(99.9),
	})
	s.NoError(err)

	// Проверим файл
	data, err := os.ReadFile(s.file)
	s.NoError(err)
	s.NotEmpty(data)

	var metrics []model.Metrics
	err = json.Unmarshal(data, &metrics)
	s.NoError(err)
	s.Equal("cpu", metrics[0].ID)
	s.InDelta(99.9, *metrics[0].Value, 1e-6)
}

func (s *TestStorageSuite) TestString_Empty() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)
	s.Empty(storage.String())
}

func (s *TestStorageSuite) TestString_NonEmpty() {
	storage, err := NewMemStorage("", 0, false, "")
	s.NoError(err)

	storage.Metrics["cpu"] = &model.Metrics{
		ID:    "cpu",
		MType: model.Gauge,
		Value: floatPtr(1.0),
	}

	str := storage.String()
	s.NotEmpty(str)

	var metrics []model.Metrics
	err = json.Unmarshal([]byte(str), &metrics)
	s.NoError(err)
	s.Equal("cpu", metrics[0].ID)
	s.InDelta(1.0, *metrics[0].Value, 1e-6)
}
