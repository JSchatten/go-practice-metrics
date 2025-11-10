package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/JSchatten/go-practice-metrics/internal/handler"
	storageService "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// help: структура для проверки JSON-ошибок
type errorResponse struct {
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func init() {
	gin.SetMode(gin.TestMode) // diable gin debug mode
}

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Valid gauge update",
			method:         http.MethodPost,
			url:            "/update/gauge/cpu_usage/99.5",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"Metric updated"}`,
		},
		{
			name:           "Valid counter update",
			method:         http.MethodPost,
			url:            "/update/counter/request_count/1",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"Metric updated"}`,
		},
		{
			name:           "Missing parameters",
			method:         http.MethodPost,
			url:            "/update/gauge//99.5",
			expectedStatus: http.StatusNotFound,
			expectedBody: errorResponse{
				StatusCode: http.StatusNotFound,
				Code:       "missing_parameters",
				Message:    "Missing required parameters",
			},
		},
		{
			name:           "Unknown metric type",
			method:         http.MethodPost,
			url:            "/update/invalid/type/99.5",
			expectedStatus: http.StatusBadRequest,
			expectedBody: errorResponse{
				StatusCode: http.StatusBadRequest,
				Code:       "unknown_metric_type",
				Message:    "Unknown metric type: invalid",
			},
		},
		{
			name:           "Invalid HTTP method",
			method:         http.MethodGet,
			url:            "/update/gauge/cpu_usage/99.5",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody: errorResponse{
				StatusCode: http.StatusMethodNotAllowed,
				Code:       "method_not_allowed",
				Message:    "Method not allowed",
			},
		},
		{
			name:           "Invalid value format for counter",
			method:         http.MethodPost,
			url:            "/update/counter/requests/abc",
			expectedStatus: http.StatusBadRequest,
			expectedBody: errorResponse{
				StatusCode: http.StatusBadRequest,
				Code:       "invalid_value_format",
				Message:    "Invalid value format: invalid counter value",
			},
		},
	}

	// Запускаем тест
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			memStorage, err := storageService.NewMemStorage(os.DevNull, 0, false, "")
			if err != nil {
				t.Fatalf("Failed to create memory storage: %v", err)
			}

			// Регистрируем маршрут
			r.POST("/update/:type/:name/:value", handler.UpdateHandler(memStorage))
			// для теста MethodNotAllowed
			r.GET("/update/:type/:name/:value", handler.UpdateHandler(memStorage))

			req, _ := http.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if body, ok := tt.expectedBody.(string); ok {
				// для простого json (успешный ответ)
				assert.JSONEq(t, body, w.Body.String())
			} else if errResp, ok := tt.expectedBody.(errorResponse); ok {
				// struct defined
				var actual errorResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				assert.Equal(t, errResp, actual)
			} else {
				t.Fatalf("unsupported expectedBody type: %T", tt.expectedBody)
			}
		})
	}

}
