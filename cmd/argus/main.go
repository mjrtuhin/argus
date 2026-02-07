package main

import (
	"context"
	"log"
	"time"

	"github.com/mjrtuhin/argus/pkg/alerting"
	"github.com/mjrtuhin/argus/pkg/detector"
	"github.com/mjrtuhin/argus/pkg/storage"
)

func main() {
	log.Println("🚀 ARGUS - Detection, Storage & Alerting")
	log.Println("=========================================")

	// Connect to database
	db, err := storage.NewDB("localhost", "5432", "argus", "argus_dev_2025", "argus")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	// Create ML client
	mlClient := detector.NewMLClient("http://localhost:5001")
	log.Println("✅ ML client created")

	// Create Slack alerting (empty webhook = console only)
	slackSender := alerting.NewSlackSender("")
	log.Println("✅ Alerting initialized (console mode)")

	ctx := context.Background()

	// Get data for metric ID 2
	metricID := 2
	since := time.Now().Add(-24 * time.Hour)

	points, err := db.GetMetricData(ctx, metricID, since)
	if err != nil {
		log.Fatalf("❌ Failed to get metric data: %v", err)
	}

	if len(points) < 10 {
		log.Printf("⚠️  Not enough data points (%d)", len(points))
		return
	}

	log.Printf("📊 Found %d data points", len(points))

	// Prepare data for ML
	var timestamps []int64
	var values []float64
	for _, p := range points {
		timestamps = append(timestamps, p.Timestamp.Unix())
		values = append(values, p.Value)
	}

	// Call ML service
	log.Println("🔮 Running ML detection...")
	result, err := mlClient.DetectAnomalies(ctx, &detector.DetectionRequest{
		MetricID:   metricID,
		MetricName: "go_gc_cycles_automatic_gc_cycles_total",
		Timestamps: timestamps,
		Values:     values,
	})
	if err != nil {
		log.Fatalf("❌ ML detection failed: %v", err)
	}

	log.Printf("✅ Detection complete: %d anomalies found", result.AnomalyCount)

	// Store and alert on anomalies
	if len(result.Anomalies) > 0 {
		log.Println("💾 Storing anomalies and sending alerts...")
		
		for _, a := range result.Anomalies {
			severity := classifySeverity(a.Score)
			
			anomaly := &storage.Anomaly{
				MetricID:         metricID,
				Timestamp:        time.Unix(a.Timestamp, 0),
				Value:            a.Value,
				AnomalyScore:     a.Score,
				DetectionMethods: a.Methods,
				Severity:         severity,
				Status:           "open",
			}

			// Store in database
			if err := db.CreateAnomaly(ctx, anomaly); err != nil {
				log.Printf("❌ Failed to store anomaly: %v", err)
				continue
			}

			log.Printf("   ✅ Stored: Value=%.2f, Score=%.3f, Severity=%s, ID=%d",
				anomaly.Value, anomaly.AnomalyScore, anomaly.Severity, anomaly.ID)

			// Send alert
			if err := slackSender.SendAlert(ctx, result.MetricName, a.Value, a.Score, severity); err != nil {
				log.Printf("❌ Failed to send alert: %v", err)
			}
		}
	} else {
		log.Println("✅ No new anomalies detected")
	}

	log.Println("\n🎉 COMPLETE: Detect → Store → Alert pipeline working!")
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
