package handler_test

import (
	"testing"

	"github.com/JSchatten/go-practice-metrics/internal/handler"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		storage storage.Storage
		want    gin.HandlerFunc
	}{
		// {
		// 	name:           "Valid Gauge",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/gauge/metric_name/123.45",
		// 	expectedStatus: http.StatusOK,
		// },
		// {
		// 	name:           "Valid Counter",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/counter/metric_name/678",
		// 	expectedStatus: http.StatusOK,
		// },
		// {
		// 	name:           "Invalid Method",
		// 	method:         http.MethodGet,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/gauge/metric_name/123.45",
		// 	expectedStatus: http.StatusMethodNotAllowed,
		// },
		// {
		// 	name:           "Invalid Content-Type",
		// 	method:         http.MethodPost,
		// 	contentType:    "application/json",
		// 	urlPath:        "/update/gauge/metric_name/123.45",
		// 	expectedStatus: http.StatusBadRequest,
		// },
		// {
		// 	name:           "Invalid URL Format",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/invalid_type/metric_name/123.45",
		// 	expectedStatus: http.StatusBadRequest,
		// },
		// {
		// 	name:           "Missing Metric Name",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/gauge//123.45",
		// 	expectedStatus: http.StatusNotFound,
		// },
		// {
		// 	name:           "Unknown Metric Type",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/unknown/metric_name/123.45",
		// 	expectedStatus: http.StatusBadRequest,
		// },
		// {
		// 	name:           "Invalid Gauge Value",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/gauge/metric_name/abc",
		// 	expectedStatus: http.StatusBadRequest,
		// },
		// {
		// 	name:           "Invalid Counter Value",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/counter/metric_name/def",
		// 	expectedStatus: http.StatusBadRequest,
		// },

		// test ya
		// {
		// 	name:           "TestGaugeHandlers/without_id",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/gauge/",
		// 	expectedStatus: http.StatusNotFound,
		// },
		// {
		// 	name:           "TestUnknownHandlers/update_invalid_type",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/unknown/testCounter/100",
		// 	expectedStatus: http.StatusOK,
		// },

		// // without id
		// {
		// 	name:           "Without Id",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update/counter//def",
		// 	expectedStatus: http.StatusBadRequest,
		// },
		// // without type
		// {
		// 	name:           "Without Id",
		// 	method:         http.MethodPost,
		// 	contentType:    "text/plain",
		// 	urlPath:        "/update//metric_name/def",
		// 	expectedStatus: http.StatusBadRequest,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.UpdateHandler(tt.storage)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("UpdateHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
