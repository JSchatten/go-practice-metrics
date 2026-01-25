// Package handler содержит HTTP-обработчики и вспомогательные функции для API.
// Включает в себя:
//   - Обработку ошибок с унифицированным форматом ответа
//   - Реализацию эндпоинтов
//   - Валидацию входных данных
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError представляет унифицированную структуру ошибки API.
//
// Используется для возврата структурированных ошибок клиенту.
//
// Поля:
//   - StatusCode: HTTP-статус ответа
//   - Code: машинно-читаемый код ошибки (например, "bad_request")
//   - Message: человеко-читаемое описание ошибки
type AppError struct {
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// Error реализует интерфейс error, возвращая текст сообщения.
func (e AppError) Error() string {
	return e.Message
}

// Errors - глобальная мапа стандартных ошибок приложения.
//
// Каждая ошибка содержит:
//   - HTTP-статус
//   - Код ошибки
//   - Сообщение по умолчанию
//
// Эти ошибки используются в хелпер-функциях вроде BodyInvalidJSON, MetricNotFound и т.д.
//
// Пример:
//
//	abortWithError(c, Errors["MetricNotFound"])
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

// abortWithError отправляет JSON-ответ с ошибкой и прерывает цепочку обработки.
// Используется всеми функциями ошибок.
func abortWithError(c *gin.Context, err AppError) {
	c.AbortWithStatusJSON(err.StatusCode, err)
}

// InvalidMetricTypeMismatch возвращает ошибку 400 с кодом "type_mismatch".
// Вызывается, когда запрошенный тип метрики не совпадает с сохранённым.
func InvalidMetricTypeMismatch(c *gin.Context) {
	abortWithError(c, Errors["MetricTypeMismatch"])
}

// BodyInvalidMetricType возвращает ошибку 400, если указан неверный тип метрики (не "gauge" и не "counter").
func BodyInvalidMetricType(c *gin.Context) {
	abortWithError(c, Errors["InvalidMetricType"])
}

// BodyMissingFields возвращает ошибку 400, если в теле запроса отсутствуют обязательные поля id или type.
func BodyMissingFields(c *gin.Context) {
	abortWithError(c, Errors["MissingRequiredFields"])
}

// BodyInvalidJSON возвращает ошибку 400, если тело запроса не является валидным JSON.
func BodyInvalidJSON(c *gin.Context) {
	abortWithError(c, Errors["InvalidJSON"])
}

// MethodNotAllowed возвращает ошибку 405, если метод не разрешён для эндпоинта.
func MethodNotAllowed(c *gin.Context) {
	abortWithError(c, Errors["MethodNotAllowed"])
}

// BadRequest возвращает общую ошибку 400.
func BadRequest(c *gin.Context) {
	abortWithError(c, Errors["BadRequest"])
}

// MissingParameters возвращает ошибку 404, если в пути отсутствуют параметры.
func MissingParameters(c *gin.Context) {
	abortWithError(c, Errors["MissingParameters"])
}

// InvalidValueFormat возвращает ошибку 400 с детализацией - например, при парсинге числа.
func InvalidValueFormat(c *gin.Context, details error) {
	err := Errors["InvalidValueFormat"]
	err.Message = err.Message + ": " + details.Error()
	abortWithError(c, err)
}

// UnknownMetricType возвращает ошибку 400 с указанием, какой тип был передан.
func UnknownMetricType(c *gin.Context, metricType string) {
	err := Errors["UnknownMetricType"]
	err.Message = err.Message + ": " + metricType
	abortWithError(c, err)
}

// MetricNotFound возвращает ошибку 404, если метрика с указанным именем не найдена.
func MetricNotFound(c *gin.Context) {
	abortWithError(c, Errors["MetricNotFound"])
}

// ValueNotProvided возвращает ошибку 400, если для gauge не указано значение.
func ValueNotProvided(c *gin.Context) {
	abortWithError(c, Errors["ValueNotProvided"])
}

// DeltaNotProvided возвращает ошибку 400, если для counter не указано значение delta.
func DeltaNotProvided(c *gin.Context) {
	abortWithError(c, Errors["DeltaNotProvided"])
}

// FailedToUpdateMetric возвращает ошибку 400 с детализацией - например, ошибка БД.
func FailedToUpdateMetric(c *gin.Context, err error) {
	appErr := Errors["FailedToUpdateMetric"]
	appErr.Message = appErr.Message + ": " + err.Error()
	abortWithError(c, appErr)
}

// BadRequestVerbose возвращает ошибку 400 с детализированным сообщением.
func BadRequestVerbose(c *gin.Context, err error) {
	appErr := Errors["BadRequestVerbose"]
	appErr.Message = appErr.Message + ": " + err.Error()
	abortWithError(c, appErr)
}

// InternalError возвращает 500 при внутренних сбоях сервера.
func InternalError(c *gin.Context) {
	abortWithError(c, Errors["InternalError"])
}
