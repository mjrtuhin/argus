package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mjrtuhin/argus/pkg/prometheus"
	"github.com/mjrtuhin/argus/pkg/storage"
)

type MetricCollector struct {
	promClient *prometheus.Client
	db         *storage.DB
	interval   time.Duration
}

func NewMetricCollector(promClient *prometheus.Client, db *storage.DB, interval time.Duration) *MetricCollector {
	return &MetricCollector{
		promClient: promClient,
		db:         db,
		interval:   interval,
	}
}

func (mc *MetricCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	log.Printf("🔄 Metric collector started (interval: %v)", mc.interval)

	// Collect immediately on start
	mc.collectMetrics(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Metric collector stopped")
			return
		case <-ticker.C:
			mc.collectMetrics(ctx)
		}
	}
}

func (mc *MetricCollector) collectMetrics(ctx context.Context) {
	// Fetch list of all metrics
	metricNames, err := mc.promClient.ListMetrics(ctx)
	if err != nil {
		log.Printf("❌ Failed to list metrics: %v", err)
		return
	}

	log.Printf("📊 Found %d metrics, collecting first 5...", len(metricNames))

	// For now, collect first 5 metrics to avoid overwhelming the system
	collected := 0
	for _, metricName := range metricNames {
		if collected >= 5 {
			break
		}

		if err := mc.collectSingleMetric(ctx, metricName); err != nil {
			log.Printf("❌ Failed to collect %s: %v", metricName, err)
			continue
		}

		collected++
	}

	log.Printf("✅ Collected %d metrics at %s", collected, time.Now().Format("15:04:05"))
}

func (mc *MetricCollector) collectSingleMetric(ctx context.Context, metricName string) error {
	// Query the metric from Prometheus
	result, err := mc.promClient.Query(ctx, metricName)
	if err != nil {
		return err
	}

	if len(result.Data.Result) == 0 {
		return nil
	}

	// Create or get metric in database
	metric, err := mc.db.CreateMetric(ctx, metricName)
	if err != nil {
		return err
	}

	// Collect all data points from this metric
	var points []storage.MetricDataPoint
	timestamp := time.Now()

	for _, r := range result.Data.Result {
		if len(r.Value) < 2 {
			continue
		}

		valueStr, ok := r.Value[1].(string)
		if !ok {
			continue
		}

		var value float64
		if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
			continue
		}

		points = append(points, storage.MetricDataPoint{
			MetricID:  metric.ID,
			Timestamp: timestamp,
			Value:     value,
		})
	}

	// Store data points
	if len(points) > 0 {
		return mc.db.InsertMetricData(ctx, points)
	}

	return nil
}
