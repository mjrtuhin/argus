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
	log.Println("🔄 Starting metric collector...")

	db, err := storage.NewDB("localhost", "5432", "argus", "argus_dev_2025", "argus")
	if err != nil {
		log.Fatalf("❌ Database error: %v", err)
	}
	defer db.Close()

	promClient := prometheus.NewClient("http://localhost:9090")
	collector := worker.NewMetricCollector(promClient, db, 60*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go collector.Start(ctx)

	log.Println("📊 Collecting metrics every 60 seconds... Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Stopped")
}
