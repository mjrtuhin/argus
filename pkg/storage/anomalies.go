package storage

import (
	"context"
	"time"

	"github.com/lib/pq"
)

type Anomaly struct {
	ID               int
	MetricID         int
	Timestamp        time.Time
	Value            float64
	AnomalyScore     float64
	DetectionMethods []string
	Severity         string
	Status           string
	RootCause        string
	Impact           string
	CreatedAt        time.Time
}

func (db *DB) CreateAnomaly(ctx context.Context, anomaly *Anomaly) error {
	return db.conn.QueryRowContext(ctx,
		`INSERT INTO anomalies 
		 (metric_id, timestamp, value, anomaly_score, detection_methods, severity, status, root_cause, impact)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at`,
		anomaly.MetricID,
		anomaly.Timestamp,
		anomaly.Value,
		anomaly.AnomalyScore,
		pq.Array(anomaly.DetectionMethods),
		anomaly.Severity,
		anomaly.Status,
		anomaly.RootCause,
		anomaly.Impact,
	).Scan(&anomaly.ID, &anomaly.CreatedAt)
}

func (db *DB) GetRecentAnomalies(ctx context.Context, limit int) ([]Anomaly, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, metric_id, timestamp, value, anomaly_score, 
		        detection_methods, severity, status, root_cause, impact, created_at
		 FROM anomalies
		 WHERE status = 'open'
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anomalies []Anomaly
	for rows.Next() {
		var a Anomaly
		if err := rows.Scan(
			&a.ID, &a.MetricID, &a.Timestamp, &a.Value,
			&a.AnomalyScore, pq.Array(&a.DetectionMethods),
			&a.Severity, &a.Status, &a.RootCause, &a.Impact, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		anomalies = append(anomalies, a)
	}

	return anomalies, rows.Err()
}

func classifySeverity(score float64) string {
	switch {
	case score >= 0.8:
		return "critical"
	case score >= 0.65:
		return "high"
	case score >= 0.5:
		return "medium"
	default:
		return "low"
	}
}
