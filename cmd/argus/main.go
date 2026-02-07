package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mjrtuhin/argus/pkg/alerting"
	"github.com/mjrtuhin/argus/pkg/api"
	"github.com/mjrtuhin/argus/pkg/detector"
	"github.com/mjrtuhin/argus/pkg/prometheus"
	"github.com/mjrtuhin/argus/pkg/storage"
	"github.com/mjrtuhin/argus/pkg/worker"
)

func main() {
	log.Println("🚀 ARGUS - Autonomous Anomaly Detection System")
	log.Println("===============================================")

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

	// Create ML client
	mlClient := detector.NewMLClient("http://localhost:5001")
	log.Println("✅ Connected to ML service")

	// Create Slack alerting
	slackSender := alerting.NewSlackSender("")
	log.Println("✅ Alerting initialized (console mode)")

	// Create API server
	apiServer := api.NewServer(db, "8080")

	// Create workers
	collector := worker.NewMetricCollector(promClient, db, 60*time.Second)
	detectorWorker := worker.NewAnomalyDetector(mlClient, db, slackSender, apiServer.GetHub(), 5*time.Minute)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start API server
	go func() {
		if err := apiServer.Start(ctx); err != nil {
			log.Printf("❌ API server error: %v", err)
		}
	}()

	// Start WebSocket hub
	go apiServer.GetHub().Run(ctx)

	// Start workers
	go collector.Start(ctx)
	go detectorWorker.Start(ctx)

	log.Println("")
	log.Println("🔄 Metric Collector: Running every 60 seconds")
	log.Println("🔮 Anomaly Detector: Running every 5 minutes")
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
