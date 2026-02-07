import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './Dashboard.css';

const API_URL = 'http://localhost:8080';

function Dashboard() {
  const [metrics, setMetrics] = useState([]);
  const [anomalies, setAnomalies] = useState([]);
  const [loading, setLoading] = useState(true);
  const [wsConnected, setWsConnected] = useState(false);

  useEffect(() => {
    fetchMetrics();
    fetchAnomalies();
    connectWebSocket();
  }, []);

  const fetchMetrics = async () => {
    try {
      const response = await axios.get(`${API_URL}/api/metrics`);
      setMetrics(response.data.metrics || []);
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
    }
  };

  const fetchAnomalies = async () => {
    try {
      const response = await axios.get(`${API_URL}/api/anomalies?limit=20`);
      console.log('Fetched anomalies:', response.data.anomalies); // Debug log
      setAnomalies(response.data.anomalies || []);
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch anomalies:', error);
      setLoading(false);
    }
  };

  const connectWebSocket = () => {
    const ws = new WebSocket(`ws://localhost:8080/ws/anomalies`);
    
    ws.onopen = () => {
      console.log('✅ WebSocket connected');
      setWsConnected(true);
    };
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'anomaly_detected') {
        console.log('📡 New anomaly:', data.anomaly);
        setAnomalies(prev => [data.anomaly, ...prev].slice(0, 20));
      }
    };
    
    ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
      setWsConnected(false);
    };
    
    ws.onclose = () => {
      console.log('🔌 WebSocket disconnected');
      setWsConnected(false);
      setTimeout(connectWebSocket, 5000);
    };
  };

  const handleResolve = async (anomalyId) => {
    try {
      setAnomalies(prev => prev.filter(a => a.id !== anomalyId));
      console.log(`Resolved anomaly ${anomalyId}`);
    } catch (error) {
      console.error('Failed to resolve anomaly:', error);
      fetchAnomalies();
    }
  };

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'critical': return '#ff0000';
      case 'high': return '#ff6b00';
      case 'medium': return '#ffcc00';
      default: return '#36a64f';
    }
  };

  const getSeverityEmoji = (severity) => {
    switch (severity) {
      case 'critical': return '🚨';
      case 'high': return '⚠️';
      case 'medium': return '⚡';
      default: return 'ℹ️';
    }
  };

  if (loading) {
    return <div className="loading">Loading ARGUS...</div>;
  }

  return (
    <div className="dashboard">
      <header className="header">
        <h1>🚀 ARGUS - Anomaly Detection</h1>
        <div className="status">
          <span className={`status-indicator ${wsConnected ? 'connected' : 'disconnected'}`}>
            {wsConnected ? '🟢 Live' : '🔴 Disconnected'}
          </span>
        </div>
      </header>

      <div className="stats">
        <div className="stat-card">
          <h3>Monitored Metrics</h3>
          <div className="stat-value">{metrics.length}</div>
        </div>
        <div className="stat-card">
          <h3>Total Anomalies</h3>
          <div className="stat-value">{anomalies.length}</div>
        </div>
        <div className="stat-card">
          <h3>Critical</h3>
          <div className="stat-value critical">
            {anomalies.filter(a => a.severity === 'critical').length}
          </div>
        </div>
        <div className="stat-card">
          <h3>High</h3>
          <div className="stat-value high">
            {anomalies.filter(a => a.severity === 'high').length}
          </div>
        </div>
      </div>

      <div className="content">
        <div className="section">
          <h2>📊 Monitored Metrics</h2>
          <div className="metrics-list">
            {metrics.map(metric => (
              <div key={metric.id} className="metric-card">
                <span className="metric-name">{metric.metric_name}</span>
                <span className={`metric-status ${metric.is_active ? 'active' : 'inactive'}`}>
                  {metric.is_active ? '✅ Active' : '⏸️ Inactive'}
                </span>
              </div>
            ))}
          </div>
        </div>

        <div className="section">
          <h2>🚨 Recent Anomalies</h2>
          <div className="anomalies-list">
            {anomalies.map(anomaly => (
              <div key={anomaly.id} className="anomaly-card" style={{borderLeftColor: getSeverityColor(anomaly.severity)}}>
                <div className="anomaly-header">
                  <span className="anomaly-emoji">{getSeverityEmoji(anomaly.severity)}</span>
                  <span className="anomaly-severity" style={{color: getSeverityColor(anomaly.severity)}}>
                    {anomaly.severity.toUpperCase()}
                  </span>
                  <span className="anomaly-time">{new Date(anomaly.timestamp).toLocaleString()}</span>
                  <button 
                    className="resolve-btn"
                    onClick={() => handleResolve(anomaly.id)}
                    title="Mark as resolved"
                  >
                    ✓ Resolve
                  </button>
                </div>
                <div className="anomaly-body">
                  <div className="anomaly-metric">{anomaly.metric_name || `Metric #${anomaly.metric_id}`}</div>
                  <div className="anomaly-details">
                    <span>Value: <strong>{anomaly.value.toFixed(2)}</strong></span>
                    <span>Score: <strong>{(anomaly.anomaly_score * 100).toFixed(1)}%</strong></span>
                  </div>
                  
                  <div className="anomaly-root-cause">
                    <strong>🔍 Root Cause:</strong> {anomaly.root_cause || 'Analyzing...'}
                  </div>
                  
                  <div className="anomaly-impact">
                    <strong>⚡ Impact:</strong> {anomaly.impact || 'Assessing...'}
                  </div>
                  
                  <div className="anomaly-methods">
                    Detected by: {anomaly.detection_methods.join(', ')}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;