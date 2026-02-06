package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mjrtuhin/argus/pkg/prometheus"
	"github.com/mjrtuhin/argus/pkg/storage"
)

func main() {
	log.Println("🚀 ARGUS - Starting...")

	// Connect to database
	db, err := storage.NewDB("localhost", "5432", "argus", "argus_dev_2025", "argus")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	// Create Prometheus client
	promClient := prometheus.NewClient("http://localhost:9090")
	log.Println("✅ Connected to Prometheus")

	// Test: Fetch one metric and store it
	ctx := context.Background()
	
	// Query a simple metric
	result, err := promClient.Query(ctx, "up")
	if err != nil {
		log.Fatalf("❌ Failed to query metric: %v", err)
	}

	if len(result.Data.Result) > 0 {
		// Create metric in database
		metric, err := db.CreateMetric(ctx, "up")
		if err != nil {
			log.Fatalf("❌ Failed to create metric: %v", err)
		}
		log.Printf("✅ Metric created in DB: %s (ID: %d)", metric.MetricName, metric.ID)

		// Store the data point
		timestamp := time.Now()
		value := result.Data.Result[0].Value[1].(string)
		
		var floatValue float64
		fmt.Sscanf(value, "%f", &floatValue)

		points := []storage.MetricDataPoint{
			{
				MetricID:  metric.ID,
				Timestamp: timestamp,
				Value:     floatValue,
			},
		}

		if err := db.InsertMetricData(ctx, points); err != nil {
			log.Fatalf("❌ Failed to insert data: %v", err)
		}
		log.Printf("✅ Data point stored: value=%f at %s", floatValue, timestamp.Format(time.RFC3339))
	}

	log.Println("🎉 ARGUS is working! Prometheus → Go → PostgreSQL pipeline complete!")
}
