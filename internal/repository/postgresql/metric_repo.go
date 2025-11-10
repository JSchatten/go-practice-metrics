package repository

import (
	"context"

	models "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricRepo struct {
	db *pgxpool.Pool
}

func NewMetricRepo(db *pgxpool.Pool) *MetricRepo {
	var metricRepo = &MetricRepo{
		db: db,
	}
	return metricRepo
}

func (r *MetricRepo) GetMetricByID(ctx context.Context, id string) (*models.Metrics, error) {
	// var m models.Metrics
	// var delta *int64
	// var value *float64

	// query := `SELECT id, type, delta, value FROM metrics WHERE id = $1`
	// err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.MType, &delta, &value)
	// if err != nil {
	// 	return nil, err
	// }

	// m.Delta = delta
	// m.Value = value

	// return &m, nil
	return nil, nil
}

func (r *MetricRepo) UpdateMetric(ctx context.Context, m *models.Metrics) error {
	// var query string
	// var args []interface{}

	// switch m.MType {
	// case "gauge":
	// 	query = `INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3)
	//              ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value`
	// 	args = []interface{}{m.ID, m.MType, m.Value}
	// case "counter":
	// 	query = `INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3)
	//              ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta`
	// 	args = []interface{}{m.ID, m.MType, m.Delta}
	// default:
	// 	return models.ErrUnknownMetricType
	// }

	// _, err := r.db.Exec(ctx, query, args...)
	// return err
	return nil
}

func (r *MetricRepo) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
