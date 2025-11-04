package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ErrBadRequest            = "Bad request"
	ErrMethodNotAllowed      = "Method not allowed"
	ErrMissingParameters     = "Missing required parameters"
	ErrInvalidValueFormat    = "Invalid value format"
	ErrUnknownMetricType     = "Unknown metric type"
	ErrMetricNotFound        = "Metric not found"
	ErrValueNotProvided      = "Value is missing for gauge"
	ErrDeltaNotProvided      = "Delta is missing for counter"
	ErrFailedToUpdateMetric  = "Failed to update metric"
	ErrInvalidJSON           = "Invalid JSON"
	ErrMissingRequiredFields = "Missing required fields: id or type"
	ErrInvalidMetricType     = "Invalid metric type: must be 'gauge' or 'counter'"
	ErrMetricTypeMissmatch   = "Metric type mismatch"
)

// Универсальная функция
func abortWithError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

// для часто используемых ошибок
func invalidMetricTypeMissmatch(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrMetricTypeMissmatch)
}

func bodyInvalidMetricType(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrInvalidMetricType)
}

func bodyMissingFields(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrMissingRequiredFields)
}

func bodyInvalidJSON(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrInvalidJSON)
}

func methodNotAllowed(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrMethodNotAllowed)
}

func badRequest(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrBadRequest)
}

func missingParameters(c *gin.Context) {
	abortWithError(c, http.StatusNotFound, ErrMissingParameters)
}

func invalidValueFormat(c *gin.Context, err error) {
	abortWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %v", ErrInvalidValueFormat, err))
}

func unknownMetricType(c *gin.Context, metricType string) {
	abortWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %s", ErrUnknownMetricType, metricType))
}

func metricNotFound(c *gin.Context) {
	abortWithError(c, http.StatusNotFound, ErrMetricNotFound)
}

func valueNotProvided(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrValueNotProvided)
}

func deltaNotProvided(c *gin.Context) {
	abortWithError(c, http.StatusBadRequest, ErrDeltaNotProvided)
}

func failedToUpdateMetric(c *gin.Context, err error) {
	abortWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %v", ErrFailedToUpdateMetric, err))
}
