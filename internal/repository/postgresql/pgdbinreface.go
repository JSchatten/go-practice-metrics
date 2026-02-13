// Package repository предоставляет интерфейс для работы с хранилищем метрик.
//
// В этом подпакете postgresql определён интерфейс MetricRepo, который описывает операции с метриками
// через PostgreSQL, включая получение, обновление и проверку соединения.
//
// Реализация интерфейса находится в файле metric_repo.go.
//
// Позволяет использовать PostgreSQL как хранилище метрик вместо файла.
package repository

// import (
// 	"context"

// 	models "github.com/JSchatten/go-practice-metrics/internal/model"
// )

// type MetricRepo interface {
// 	GetMetricByID(ctx context.Context, id string) (*models.Metrics, error)
// 	UpdateMetric(ctx context.Context, m *models.Metrics) error
// 	Ping(ctx context.Context) error
// }
