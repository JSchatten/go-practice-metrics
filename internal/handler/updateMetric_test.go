package handler_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/JSchatten/go-practice-metrics/internal/handler"
	storageService "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
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
			expectedBody:   `{"error":"Missing required parameters"}`,
		},
		{
			name:           "Unknown metric type",
			method:         http.MethodPost,
			url:            "/update/invalid/type/99.5",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Unknown metric type: invalid"}`,
		},
		// { // 404 page not found - returns by gin deafult, maybe middleware after next()
		// 	name:           "Invalid HTTP method",
		// 	method:         http.MethodGet,
		// 	url:            "/update/gauge/cpu_usage/99.5",
		// 	expectedStatus: http.StatusNotFound,
		// 	expectedBody:   `{"error":"Method not allowed"}`,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем инстанс Gin и регистрируем обработчик
			r := gin.Default()
			r.POST("/update/:type/:name/:value", handler.UpdateHandler(storageService.NewMemStorage(os.DevNull, 0, false)))

			// Создаем запрос
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Проверяем статус и тело ответа
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}

}
