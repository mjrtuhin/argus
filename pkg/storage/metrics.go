package storage

import (
	"context"
	"time"
)

type Metric struct {
	ID              int
	MetricName      string
	IsActive        bool
	LastCollectedAt *time.Time
}

type MetricDataPoint struct {
	MetricID  int
	Timestamp time.Time
	Value     float64
}

func (db *DB) CreateMetric(ctx context.Context, metricName string) (*Metric, error) {
	var metric Metric
	err := db.conn.QueryRowContext(ctx,
		`INSERT INTO metrics (metric_name, is_active) 
		 VALUES ($1, true) 
		 ON CONFLICT (metric_name) DO UPDATE SET is_active = true
		 RETURNING id, metric_name, is_active`,
		metricName,
	).Scan(&metric.ID, &metric.MetricName, &metric.IsActive)

	return &metric, err
}

func (db *DB) InsertMetricData(ctx context.Context, points []MetricDataPoint) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO metric_data (metric_id, timestamp, value) 
		 VALUES ($1, $2, $3) 
		 ON CONFLICT (metric_id, timestamp) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, point := range points {
		_, err := stmt.ExecContext(ctx, point.MetricID, point.Timestamp, point.Value)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetMetrics(ctx context.Context) ([]Metric, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, metric_name, is_active, last_collected_at 
		 FROM metrics 
		 WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		if err := rows.Scan(&m.ID, &m.MetricName, &m.IsActive, &m.LastCollectedAt); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}
