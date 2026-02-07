package worker

import (
	"context"
	"log"
	"time"

	"github.com/mjrtuhin/argus/pkg/alerting"
	"github.com/mjrtuhin/argus/pkg/detector"
	"github.com/mjrtuhin/argus/pkg/storage"
)

type AnomalyDetector struct {
	mlClient    *detector.MLClient
	db          *storage.DB
	slackSender *alerting.SlackSender
	interval    time.Duration
}

func NewAnomalyDetector(mlClient *detector.MLClient, db *storage.DB, slackSender *alerting.SlackSender, interval time.Duration) *AnomalyDetector {
	return &AnomalyDetector{
		mlClient:    mlClient,
		db:          db,
		slackSender: slackSender,
		interval:    interval,
	}
}

func (ad *AnomalyDetector) Start(ctx context.Context) {
	ticker := time.NewTicker(ad.interval)
	defer ticker.Stop()

	log.Printf("🔮 Anomaly detector started (interval: %v)", ad.interval)

	// Run immediately on start
	ad.runDetection(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Anomaly detector stopped")
			return
		case <-ticker.C:
			ad.runDetection(ctx)
		}
	}
}

func (ad *AnomalyDetector) runDetection(ctx context.Context) {
	// Get all active metrics
	metrics, err := ad.db.GetMetrics(ctx)
	if err != nil {
		log.Printf("❌ Failed to get metrics: %v", err)
		return
	}

	log.Printf("🔍 Running detection on %d metrics...", len(metrics))

	detectedCount := 0
	for _, metric := range metrics {
		count, err := ad.detectForMetric(ctx, metric)
		if err != nil {
			log.Printf("❌ Detection failed for metric %s: %v", metric.MetricName, err)
			continue
		}
		detectedCount += count
	}

	log.Printf("✅ Detection complete: %d new anomalies found at %s",
		detectedCount, time.Now().Format("15:04:05"))
}

func (ad *AnomalyDetector) detectForMetric(ctx context.Context, metric storage.Metric) (int, error) {
	// Get data from last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	points, err := ad.db.GetMetricData(ctx, metric.ID, since)
	if err != nil {
		return 0, err
	}

	// Need at least 10 points
	if len(points) < 10 {
		return 0, nil
	}

	// Prepare data for ML
	var timestamps []int64
	var values []float64
	for _, p := range points {
		timestamps = append(timestamps, p.Timestamp.Unix())
		values = append(values, p.Value)
	}

	// Call ML service
	result, err := ad.mlClient.DetectAnomalies(ctx, &detector.DetectionRequest{
		MetricID:   metric.ID,
		MetricName: metric.MetricName,
		Timestamps: timestamps,
		Values:     values,
	})
	if err != nil {
		return 0, err
	}

	// Store and alert on new anomalies
	newAnomalies := 0
	for _, a := range result.Anomalies {
		severity := classifySeverity(a.Score)

		anomaly := &storage.Anomaly{
			MetricID:         metric.ID,
			Timestamp:        time.Unix(a.Timestamp, 0),
			Value:            a.Value,
			AnomalyScore:     a.Score,
			DetectionMethods: a.Methods,
			Severity:         severity,
			Status:           "open",
		}

		// Store in database
		if err := ad.db.CreateAnomaly(ctx, anomaly); err != nil {
			// Might be duplicate - skip
			continue
		}

		newAnomalies++

		// Send alert
		if err := ad.slackSender.SendAlert(ctx, metric.MetricName, a.Value, a.Score, severity); err != nil {
			log.Printf("⚠️  Failed to send alert: %v", err)
		}
	}

	return newAnomalies, nil
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
