package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	models "github.com/JSchatten/go-practice-metrics/internal/model"
	"github.com/golang-migrate/migrate/v4"
	pgxMigrate "github.com/golang-migrate/migrate/v4/database/pgx"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // активация драйвера дл миграции
	"github.com/rs/zerolog/log"
)

type MetricRepo struct {
	db       *pgxpool.Pool
	dsn      string
	pgConfig *pgxpool.Config
}

func NewMetricRepo(dsn string) (*MetricRepo, error) {

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	var metricRepo = &MetricRepo{
		db:       pool,
		dsn:      dsn,
		pgConfig: config,
	}
	return metricRepo, nil
}

func (r *MetricRepo) Migrate(ctx context.Context) error {
	if err := r.db.Ping(ctx); err != nil {
		log.Logger.Error().Err(err).Msg("Error init migration driver")
		return err
	}

	db, err := sql.Open("pgx", r.dsn)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to open database for migration")
		return err
	}
	defer db.Close()

	// Проверим соединение
	if err := db.Ping(); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to ping database")
		return err
	}

	// Создаём экземпляр драйвера миграций
	driver, err := pgxMigrate.WithInstance(db, &pgxMigrate.Config{})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to create migrate driver instance")
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // путь к миграциям
		"pgx",               // имя драйвера (должно совпадать с тем, что в WithInstance)
		driver,
	)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to create migrate instance")
		return err
	}

	// Выполняем миграции
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Logger.Error().Err(err).Msg("Migration failed")
		return err
	}
	return nil
}

func (r *MetricRepo) GetMetricByID(ctx context.Context, id string) (*models.Metrics, error) {
	var m models.Metrics
	var delta *int64
	var value *float64

	query := `SELECT id, type, delta, value FROM metrics WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.MType, &delta, &value)
	if err != nil {
		return nil, err
	}

	m.Delta = delta
	m.Value = value

	return &m, nil
}

func (r *MetricRepo) UpdateMetric(ctx context.Context, m *models.Metrics) error {
	var query string
	var args []interface{}

	switch m.MType {
	case "gauge":
		query = `INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3)
	             ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value`
		args = []interface{}{m.ID, m.MType, m.Value}
	case "counter":
		query = `INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3)
	             ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta`
		args = []interface{}{m.ID, m.MType, m.Delta}
	default:
		return models.ErrUnknownMetricType
	}

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *MetricRepo) UpdateMetricBatch(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	// Разделяем метрики
	var gauges []models.Metrics
	var counters []models.Metrics

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value != nil {
				gauges = append(gauges, m)
			}
		case "counter":
			if m.Delta != nil {
				counters = append(counters, m)
			}
		default:
			return models.ErrUnknownMetricType
		}
	}

	// Обрабатываем batch
	return r.execBatchUpsert(ctx, gauges, counters)

}

func (r *MetricRepo) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *MetricRepo) execBatchUpsert(ctx context.Context, gauges, counters []models.Metrics) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	rawConn := conn.Conn()

	// Подготавливаем и выполняем gauge, если есть
	if len(gauges) > 0 {
		if err := r.prepareAndUpsertGauges(ctx, conn.Conn(), gauges); err != nil {
			return err
		}
	}

	// Подготавливаем и выполняем counter, если есть
	if len(counters) > 0 {
		if err := r.prepareAndUpsertCounters(ctx, rawConn, counters); err != nil {
			return err
		}
	}

	return nil
}

func (r *MetricRepo) prepareAndUpsertGauges(ctx context.Context, conn *pgx.Conn, gauges []models.Metrics) error {
	const stmtName = "upsert_gauge_batch"

	// Подготавливаем запрос — если уже подготовлен, pgx ничего не сделает (PostgreSQL вернёт ошибку дубликата, но pgx её игнорирует)
	_, err := conn.Prepare(ctx, stmtName, `
		INSERT INTO metrics (id, type, value) 
		VALUES (unnest($1::text[]), 'gauge', unnest($2::double precision[]))
		ON CONFLICT (id) DO UPDATE 
		SET value = EXCLUDED.value
	`)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to prepare statement %s: %w", stmtName, err)
	}

	// Собираем данные
	ids := make([]string, len(gauges))
	values := make([]float64, len(gauges))
	for i, g := range gauges {
		ids[i] = g.ID
		values[i] = *g.Value
	}

	// Выполняем
	_, err = conn.Exec(ctx, stmtName, ids, values)
	return err
}

func (r *MetricRepo) prepareAndUpsertCounters(ctx context.Context, conn *pgx.Conn, counters []models.Metrics) error {
	const stmtName = "upsert_counter_batch"

	_, err := conn.Prepare(ctx, stmtName, `
		INSERT INTO metrics (id, type, delta) 
		VALUES (unnest($1::text[]), 'counter', unnest($2::bigint[]))
		ON CONFLICT (id) DO UPDATE 
		SET delta = metrics.delta + EXCLUDED.delta
	`)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to prepare statement %s: %w", stmtName, err)
	}

	// Собираем данные
	ids := make([]string, len(counters))
	deltas := make([]int64, len(counters))
	for i, c := range counters {
		ids[i] = c.ID
		deltas[i] = *c.Delta
	}

	// Выполняем
	_, err = conn.Exec(ctx, stmtName, ids, deltas)
	return err
}
