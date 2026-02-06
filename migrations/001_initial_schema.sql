-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- METRICS TABLE
CREATE TABLE metrics (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(255) NOT NULL UNIQUE,
    labels JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_collected_at TIMESTAMPTZ
);

CREATE INDEX idx_metrics_active ON metrics(is_active);

-- METRIC_DATA TABLE (time-series)
CREATE TABLE metric_data (
    metric_id INT NOT NULL REFERENCES metrics(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (metric_id, timestamp)
);

-- Convert to TimescaleDB hypertable
SELECT create_hypertable('metric_data', 'timestamp');

-- ANOMALIES TABLE
CREATE TABLE anomalies (
    id SERIAL PRIMARY KEY,
    metric_id INT NOT NULL REFERENCES metrics(id),
    timestamp TIMESTAMPTZ NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    anomaly_score DOUBLE PRECISION NOT NULL,
    detection_methods TEXT[],
    severity VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_anomalies_metric ON anomalies(metric_id);
CREATE INDEX idx_anomalies_status ON anomalies(status);
