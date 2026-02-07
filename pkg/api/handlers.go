package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

type MetricsResponse struct {
	Metrics []MetricInfo `json:"metrics"`
	Total   int          `json:"total"`
}

type MetricInfo struct {
	ID              int    `json:"id"`
	MetricName      string `json:"metric_name"`
	IsActive        bool   `json:"is_active"`
	LastCollectedAt string `json:"last_collected_at,omitempty"`
}

type AnomaliesResponse struct {
	Anomalies []AnomalyInfo `json:"anomalies"`
	Total     int           `json:"total"`
}

type AnomalyInfo struct {
	ID               int      `json:"id"`
	MetricID         int      `json:"metric_id"`
	MetricName       string   `json:"metric_name,omitempty"`
	Timestamp        string   `json:"timestamp"`
	Value            float64  `json:"value"`
	AnomalyScore     float64  `json:"anomaly_score"`
	DetectionMethods []string `json:"detection_methods"`
	Severity         string   `json:"severity"`
	Status           string   `json:"status"`
	RootCause        string   `json:"root_cause"`
	Impact           string   `json:"impact"`
	CreatedAt        string   `json:"created_at"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "healthy",
		Service: "argus-api",
		Time:    timeNow(),
	}
	respondJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := s.db.GetMetrics(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch metrics")
		return
	}

	metricInfos := make([]MetricInfo, len(metrics))
	for i, m := range metrics {
		metricInfos[i] = MetricInfo{
			ID:         m.ID,
			MetricName: m.MetricName,
			IsActive:   m.IsActive,
		}
		if m.LastCollectedAt != nil {
			metricInfos[i].LastCollectedAt = m.LastCollectedAt.Format("2006-01-02T15:04:05Z")
		}
	}

	response := MetricsResponse{
		Metrics: metricInfos,
		Total:   len(metricInfos),
	}
	respondJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetAnomalies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get limit from query params (default 50)
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	anomalies, err := s.db.GetRecentAnomalies(ctx, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch anomalies")
		return
	}

	anomalyInfos := make([]AnomalyInfo, len(anomalies))
	for i, a := range anomalies {
		anomalyInfos[i] = AnomalyInfo{
			ID:               a.ID,
			MetricID:         a.MetricID,
			Timestamp:        a.Timestamp.Format("2006-01-02T15:04:05Z"),
			Value:            a.Value,
			AnomalyScore:     a.AnomalyScore,
			DetectionMethods: a.DetectionMethods,
			Severity:         a.Severity,
			Status:           a.Status,
			RootCause:        a.RootCause,
			Impact:           a.Impact,
			CreatedAt:        a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	response := AnomaliesResponse{
		Anomalies: anomalyInfos,
		Total:     len(anomalyInfos),
	}
	respondJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetAnomalyByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid anomaly ID")
		return
	}

	ctx := r.Context()
	anomalies, err := s.db.GetRecentAnomalies(ctx, 100)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch anomaly")
		return
	}

	for _, a := range anomalies {
		if a.ID == id {
			anomalyInfo := AnomalyInfo{
				ID:               a.ID,
				MetricID:         a.MetricID,
				Timestamp:        a.Timestamp.Format("2006-01-02T15:04:05Z"),
				Value:            a.Value,
				AnomalyScore:     a.AnomalyScore,
				DetectionMethods: a.DetectionMethods,
				Severity:         a.Severity,
				Status:           a.Status,
				RootCause:        a.RootCause,
				Impact:           a.Impact,
				CreatedAt:        a.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
			respondJSON(w, http.StatusOK, anomalyInfo)
			return
		}
	}

	respondError(w, http.StatusNotFound, "Anomaly not found")
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func timeNow() string {
	return time.Now().Format("2006-01-02T15:04:05Z")
}