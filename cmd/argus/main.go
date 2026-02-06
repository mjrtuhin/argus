package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mjrtuhin/argus/pkg/prometheus"
	"github.com/mjrtuhin/argus/pkg/storage"
	"github.com/mjrtuhin/argus/pkg/worker"
)

func main() {
	log.Println("🚀 ARGUS - ML-Powered Anomaly Detection System")
	log.Println("=" + string(make([]byte, 50)))

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

	// Create metric collector (collect every 60 seconds)
	collector := worker.NewMetricCollector(promClient, db, 60*time.Second)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collector in background
	go collector.Start(ctx)

	log.Println("🔄 Metric collection started (every 60 seconds)")
	log.Println("📊 Press Ctrl+C to stop")
	log.Println("")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("\n🛑 Shutting down gracefully...")
	cancel()
	time.Sleep(2 * time.Second)
	log.Println("👋 Goodbye!")
}
