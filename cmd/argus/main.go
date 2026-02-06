package main

import (
	"context"
	"log"
	"time"

	"github.com/mjrtuhin/argus/pkg/detector"
	"github.com/mjrtuhin/argus/pkg/storage"
)

func main() {
	log.Println("🚀 ARGUS - Testing ML Detection Pipeline")
	log.Println("========================================")

	// Connect to database
	db, err := storage.NewDB("localhost", "5432", "argus", "argus_dev_2025", "argus")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	// Create ML client (port 5001)
	mlClient := detector.NewMLClient("http://localhost:5001")
	log.Println("✅ ML client created")

	ctx := context.Background()

	// Get metric ID 2 (first collected metric)
	metricID := 2
	since := time.Now().Add(-24 * time.Hour)

	// Fetch data from database
	points, err := db.GetMetricData(ctx, metricID, since)
	if err != nil {
		log.Fatalf("❌ Failed to get metric data: %v", err)
	}

	if len(points) < 10 {
		log.Printf("⚠️  Not enough data points (%d), need at least 10", len(points))
		log.Println("Run the collector for a few minutes first!")
		return
	}

	log.Printf("📊 Found %d data points for metric ID %d", len(points), metricID)

	// Prepare data for ML
	var timestamps []int64
	var values []float64
	for _, p := range points {
		timestamps = append(timestamps, p.Timestamp.Unix())
		values = append(values, p.Value)
	}

	// Call ML service
	log.Println("🔮 Sending data to ML service...")
	result, err := mlClient.DetectAnomalies(ctx, &detector.DetectionRequest{
		MetricID:   metricID,
		MetricName: "test_metric",
		Timestamps: timestamps,
		Values:     values,
	})
	if err != nil {
		log.Fatalf("❌ ML detection failed: %v", err)
	}

	// Display results
	log.Printf("✅ ML Analysis Complete!")
	log.Printf("   Total Points: %d", result.TotalPoints)
	log.Printf("   Anomalies Found: %d", result.AnomalyCount)

	if len(result.Anomalies) > 0 {
		log.Println("\n🚨 ANOMALIES DETECTED:")
		for i, a := range result.Anomalies {
			ts := time.Unix(a.Timestamp, 0)
			log.Printf("   %d. Value: %.2f | Score: %.3f | Time: %s", 
				i+1, a.Value, a.Score, ts.Format("15:04:05"))
		}
	} else {
		log.Println("\n✅ No anomalies detected - all metrics are normal!")
	}

	log.Println("\n🎉 PIPELINE TEST COMPLETE: Go → PostgreSQL → Python ML → Results")
}
