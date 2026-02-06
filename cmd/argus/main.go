package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mjrtuhin/argus/pkg/prometheus"
)

func main() {
	log.Println("🚀 ARGUS - Starting...")

	// Create Prometheus client
	promClient := prometheus.NewClient("http://localhost:9090")

	// Test: List all available metrics
	ctx := context.Background()
	metrics, err := promClient.ListMetrics(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to fetch metrics: %v", err)
	}

	fmt.Printf("✅ Found %d metrics in Prometheus\n", len(metrics))
	fmt.Println("\nFirst 10 metrics:")
	for i, metric := range metrics {
		if i >= 10 {
			break
		}
		fmt.Printf("  %d. %s\n", i+1, metric)
	}
}
