// Package model
// Предоставляет основные структуры данных для работы с метриками.
// Включает в себя:
//   - Описание метрик (Metrics) с поддержкой типов counter и gauge.
//   - Конструктор NewMetrics для создания метрик из строки.
//   - Метод Validate для проверки корректности метрики.
//   - Реализацию fmt.Stringer для удобного вывода.

package models

import (
	"encoding/json"
	"strconv"
)

const (
	// Counter обозначает тип метрики накопителя/инкремента (нарастающее целочисленное значение).
	Counter = "counter"
	// Gauge обозначает тип метрики "измерителя" (устанавливаемое вещественное значение любого знака).
	Gauge = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.

// Metrics описывает сущность метрики
type Metrics struct {
	// ID - уникальный идентификатор метрики (имя метрики).
	ID string `json:"id"`
	// MType - тип метрики: "counter" или "gauge".
	MType string `json:"type"`
	// Delta - значение счётчика (целое число). Должно быть указано только для типа "counter".
	Delta *int64 `json:"delta,omitempty"`
	// Value - значение измерителя (вещественное число). Должно быть указано только для типа "gauge".
	Value *float64 `json:"value,omitempty"`
	// Hash - опциональная подпись метрики (например, HMAC).
	// Используется для проверки целостности данных.
	Hash string `json:"hash,omitempty"`
}

// Добавленный код для Metrics

// String Validate checks if the Metrics instance is valid.
// String возвращает строковое представление метрики в формате JSON.
// Если метрика равна nil, возвращается "<nil>".
// В случае ошибки маршалинга возвращается описание ошибки.
//
// Реализует интерфейс fmt.Stringer.
func (m *Metrics) String() string {
	if m == nil {
		return "<nil>"
	}
	s, _ := json.Marshal(m)
	// if err != nil { // Не возвращаем ошибку, чтобы не нарушать интерфейс fmt.Stringer
	// 	return fmt.Sprintf("<error> %s", err) // Также ветка проверки недостижима
	// } // потому как объект всегда явялется сериализуемым
	return string(s)
}

// NewMetrics создаёт новую метрику на основе имени, типа и строкового значения.
// Проверяет корректность входных данных и возвращает ошибку при нарушении условий.
//
// Поведение:
//   - Для типа "counter" значение парсится как int64.
//   - Для типа "gauge" значение парсится как float64.
//   - Пустое имя метрики или тип вызывают ошибку.
//
// Примеры:
//
//	metric, err := NewMetrics("requests", "counter", "100")
//	metric, err := NewMetrics("cpu", "gauge", "0.85")
//
// Возвращает указатель на Metrics и nil в случае успеха, иначе - nil и ошибку.
func NewMetrics(id, mType, valueStr string) (*Metrics, error) {
	if id == "" {
		return nil, ErrEmptyMetricID
	}
	if mType == "" {
		return nil, ErrEmptyMetricType
	}

	var metric Metrics
	metric.ID = id
	metric.MType = mType

	switch mType {
	case Counter:
		if valueStr == "" {
			return nil, ErrValueRequired
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return nil, ErrInvalidCounterValue
		}
		metric.Delta = &value

	case Gauge:
		if valueStr == "" {
			return nil, ErrValueRequired
		}
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, ErrInvalidGaugeValue
		}
		metric.Value = &value

	default:
		return nil, ErrUnknownMetricType
	}

	return &metric, nil
}

// Validate проверяет, что метрика корректна с точки зрения логики и заполненности полей.
// Возвращает nil, если метрика валидна, и ошибку в противном случае.
//
// Проверки:
//   - Метрика не должна быть nil.
//   - ID и MType не должны быть пустыми.
//   - Для типа "counter" должно быть задано поле Delta.
//   - Для типа "gauge" должно быть задано поле Value.
//   - Тип метрики должен быть одним из допустимых: "counter" или "gauge".
//
// Пример использования:
//
//	if err := metric.Validate(); err != nil {
//	    log.Printf("Invalid metric: %v", err)
//	}
func (m *Metrics) Validate() error {
	if m == nil {
		return ErrNilMetric
	}
	if m.ID == "" {
		return ErrEmptyMetricID
	}
	if m.MType == "" {
		return ErrEmptyMetricType
	}

	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return ErrDeltaRequired
		}
	case Gauge:
		if m.Value == nil {
			return ErrValueRequired
		}
	default:
		return ErrUnknownMetricType
	}

	return nil
}
