package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handler "github.com/JSchatten/go-practice-metrics/internal/handler"
	model "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

// MockStorage реализует интерфейс Storage для тестирования
type MockStorage struct {
	UpdateMetricCalls []*model.Metrics
}

func (m *MockStorage) UpdateMetric(metric *model.Metrics) error {
	m.UpdateMetricCalls = append(m.UpdateMetricCalls, metric)
	return nil
}

func (m *MockStorage) GetMetric(id string) *model.Metrics {
	return nil
}

// func newMockStorage() *MockStorage {
// 	return &MockStorage{
// 		UpdateMetricCalls: []*model.Metrics{},
// 	}
// }

// Тест для LiveHandler
func TestLiveHandler(t *testing.T) {
	handler := handler.LiveHandler()
	req := httptest.NewRequest("GET", "/live/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := strings.TrimSpace(w.Body.String())
	assert.Equal(t, "It's alive!", body)
}

func TestUpdateHandler(t *testing.T) {
	mockStorage := &MockStorage{}

	handler := handler.UpdateHandler(mockStorage)

	tests := []struct {
		name           string
		method         string
		contentType    string
		urlPath        string
		expectedStatus int
	}{
		{
			name:           "Valid Gauge",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/gauge/metric_name/123.45",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid Counter",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/counter/metric_name/678",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			contentType:    "text/plain",
			urlPath:        "/update/gauge/metric_name/123.45",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid Content-Type",
			method:         http.MethodPost,
			contentType:    "application/json",
			urlPath:        "/update/gauge/metric_name/123.45",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid URL Format",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/invalid_type/metric_name/123.45",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Metric Name",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/gauge//123.45",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unknown Metric Type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/unknown/metric_name/123.45",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Gauge Value",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/gauge/metric_name/abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Counter Value",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlPath:        "/update/counter/metric_name/def",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Очистка перед каждым тестом
			mockStorage.UpdateMetricCalls = []*model.Metrics{}

			req := httptest.NewRequest(tt.method, tt.urlPath, nil)
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Проверка вызова UpdateMetric только для успешных тестов
			if tt.expectedStatus == http.StatusOK {
				assert.Len(t, mockStorage.UpdateMetricCalls, 1)
				metric := mockStorage.UpdateMetricCalls[0]
				assert.Equal(t, "metric_name", metric.ID)
				switch tt.urlPath {
				case "/update/gauge/metric_name/123.45":
					assert.Equal(t, model.Gauge, metric.MType)
					assert.Equal(t, float64(123.45), *metric.Value)
					assert.Nil(t, metric.Delta)
				case "/update/counter/metric_name/678":
					assert.Equal(t, model.Counter, metric.MType)
					assert.Nil(t, metric.Value)
					assert.Equal(t, int64(678), *metric.Delta)
				}
			} else {
				assert.Len(t, mockStorage.UpdateMetricCalls, 0)
			}
		})
	}

}
