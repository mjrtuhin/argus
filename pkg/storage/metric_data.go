package storage

import (
	"context"
	"time"
)

func (db *DB) GetMetricData(ctx context.Context, metricID int, since time.Time) ([]MetricDataPoint, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT metric_id, timestamp, value 
		 FROM metric_data 
		 WHERE metric_id = $1 AND timestamp >= $2
		 ORDER BY timestamp ASC`,
		metricID, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []MetricDataPoint
	for rows.Next() {
		var p MetricDataPoint
		if err := rows.Scan(&p.MetricID, &p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, rows.Err()
}
