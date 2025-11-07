package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// Error реализует интерфейс error
func (e AppError) Error() string {
	return e.Message
}

var Errors = map[string]AppError{
	"BadRequest":            {http.StatusBadRequest, "bad_request", "Bad request"},
	"MethodNotAllowed":      {http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed"},
	"MissingParameters":     {http.StatusNotFound, "missing_parameters", "Missing required parameters"},
	"InvalidValueFormat":    {http.StatusBadRequest, "invalid_value_format", "Invalid value format"},
	"UnknownMetricType":     {http.StatusBadRequest, "unknown_metric_type", "Unknown metric type"},
	"MetricNotFound":        {http.StatusNotFound, "metric_not_found", "Metric not found"},
	"ValueNotProvided":      {http.StatusBadRequest, "value_not_provided", "Value is missing for gauge"},
	"DeltaNotProvided":      {http.StatusBadRequest, "delta_not_provided", "Delta is missing for counter"},
	"FailedToUpdateMetric":  {http.StatusBadRequest, "update_failed", "Failed to update metric"},
	"InvalidJSON":           {http.StatusBadRequest, "invalid_json", "Invalid JSON"},
	"MissingRequiredFields": {http.StatusBadRequest, "missing_fields", "Missing required fields: id or type"},
	"InvalidMetricType":     {http.StatusBadRequest, "invalid_metric_type", "Invalid metric type: must be 'gauge' or 'counter'"},
	"MetricTypeMismatch":    {http.StatusBadRequest, "type_mismatch", "Metric type mismatch"},
	"InternalError":         {http.StatusInternalServerError, "internal_error", "Server internal error"},
	"BadRequestVerbose":     {http.StatusBadRequest, "bad_request", "Bad request"},
}

func abortWithError(c *gin.Context, err AppError) {
	c.AbortWithStatusJSON(err.StatusCode, err)
}

// Пример: вместо invalidMetricTypeMissmatch(c)
func InvalidMetricTypeMismatch(c *gin.Context) {
	abortWithError(c, Errors["MetricTypeMismatch"])
}

func BodyInvalidMetricType(c *gin.Context) {
	abortWithError(c, Errors["InvalidMetricType"])
}

func BodyMissingFields(c *gin.Context) {
	abortWithError(c, Errors["MissingRequiredFields"])
}

func BodyInvalidJSON(c *gin.Context) {
	abortWithError(c, Errors["InvalidJSON"])
}

func MethodNotAllowed(c *gin.Context) {
	abortWithError(c, Errors["MethodNotAllowed"])
}

func BadRequest(c *gin.Context) {
	abortWithError(c, Errors["BadRequest"])
}

func MissingParameters(c *gin.Context) {
	abortWithError(c, Errors["MissingParameters"])
}

func InvalidValueFormat(c *gin.Context, details error) {
	err := Errors["InvalidValueFormat"]
	err.Message = err.Message + ": " + details.Error()
	abortWithError(c, err)
}

func UnknownMetricType(c *gin.Context, metricType string) {
	err := Errors["UnknownMetricType"]
	err.Message = err.Message + ": " + metricType
	abortWithError(c, err)
}

func MetricNotFound(c *gin.Context) {
	abortWithError(c, Errors["MetricNotFound"])
}

func ValueNotProvided(c *gin.Context) {
	abortWithError(c, Errors["ValueNotProvided"])
}

func DeltaNotProvided(c *gin.Context) {
	abortWithError(c, Errors["DeltaNotProvided"])
}

func FailedToUpdateMetric(c *gin.Context, err error) {
	appErr := Errors["FailedToUpdateMetric"]
	appErr.Message = appErr.Message + ": " + err.Error()
	abortWithError(c, appErr)
}

func BadRequestVerbose(c *gin.Context, err error) {
	appErr := Errors["BadRequestVerbose"]
	appErr.Message = appErr.Message + ": " + err.Error()
	abortWithError(c, appErr)
}

func InternalError(c *gin.Context) {
	abortWithError(c, Errors["InternalError"])
}
