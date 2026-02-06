package detector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

type DetectionRequest struct {
	MetricID   int       `json:"metric_id"`
	MetricName string    `json:"metric_name"`
	Timestamps []int64   `json:"timestamps"`
	Values     []float64 `json:"values"`
}

type Anomaly struct {
	Timestamp int64    `json:"timestamp"`
	Value     float64  `json:"value"`
	Score     float64  `json:"score"`
	Methods   []string `json:"methods"`
}

type DetectionResponse struct {
	MetricID      int       `json:"metric_id"`
	MetricName    string    `json:"metric_name"`
	Anomalies     []Anomaly `json:"anomalies"`
	TotalPoints   int       `json:"total_points"`
	AnomalyCount  int       `json:"anomaly_count"`
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *MLClient) DetectAnomalies(ctx context.Context, req *DetectionRequest) (*DetectionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/detect", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d", resp.StatusCode)
	}

	var result DetectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
