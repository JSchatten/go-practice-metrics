// Package repository
// определяет интерфейсы для хранения и сериализации метрик.
//
// В целях соблюдения принципа единственной ответственности
// интерфейс MetricRepository был разделён на два:
//
// MetricStorage - отвечает за прямое управление метриками:
//   - Получение метрики по ID
//   - Обновление значения метрики
//   - Проверка доступности хранилища
//
// MetricSerializer - отвечает за работу с сериализованными данными:
//   - Сохранение всех метрик в виде байт
//   - Загрузка всех метрик из байт
//
// Это позволяет независимо реализовывать и тестировать
// логику хранения и сериализации.
package repository

import (
	"context"

	models "github.com/JSchatten/go-practice-metrics/internal/model"
)

// MetricStorage - интерфейс для прямого доступа к метрикам.
// Реализуется хранилищами, такими как память, файл или PostgreSQL.
// Гарантирует базовые операции чтения и записи отдельных метрик.
type MetricStorage interface {
	// GetMetricByID возвращает метрику по её идентификатору.
	// Возвращается указатель на метрику и ошибка (например, ErrNotFound).
	GetMetricByID(ctx context.Context, id string) (*models.Metrics, error)

	// UpdateMetric обновляет значение метрики.
	// Для счётчиков (counter) значение накапливается, для измерителей (gauge) - заменяется.
	UpdateMetric(ctx context.Context, m *models.Metrics) error

	// Ping проверяет доступность хранилища.
	// Используется для health-check.
	Ping(ctx context.Context) error
}

// MetricSerializer - интерфейс для сериализации и десериализации метрик.
// Отвечает за сохранение и загрузку состояния всех метрик в/из байтового потока.
// Может быть реализован отдельно от хранилища (например, через JSON, Gob и т.п.).
type MetricSerializer interface {
	// SaveMetrics сохраняет сериализованные метрики в постоянное хранилище.
	// Должен быть потокобезопасным. Возвращает ошибку при неудаче.
	SaveMetrics(data []byte) error

	// LoadMetrics загружает сериализованные метрики из постоянного хранилища.
	// Возвращает данные и ошибку (если данные отсутствуют, может вернуть nil, nil).
	LoadMetrics() ([]byte, error)
}

// MetricRepository - УСТАРЕЛ.
// Был разделён на MetricStorage и MetricSerializer
// для соблюдения принципа единственной ответственности.
type MetricRepository interface {
	MetricStorage
	MetricSerializer
}
