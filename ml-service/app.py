from flask import Flask, request, jsonify
import numpy as np
import pandas as pd
from prophet import Prophet
from statsmodels.tsa.seasonal import STL
from sklearn.ensemble import IsolationForest
import logging
import warnings

warnings.filterwarnings('ignore')

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)

@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'healthy', 'service': 'argus-ml'})

@app.route('/detect', methods=['POST'])
def detect_anomalies():
    try:
        data = request.json
        
        metric_id = data.get('metric_id')
        metric_name = data.get('metric_name')
        timestamps = data.get('timestamps', [])
        values = data.get('values', [])
        
        if len(values) < 20:
            return jsonify({
                'error': 'Need at least 20 data points for ensemble detection',
                'anomalies': []
            }), 400
        
        # Run ensemble detection
        anomalies = detect_ensemble(timestamps, values)
        
        logging.info(f"✅ Detected {len(anomalies)} anomalies for {metric_name}")
        
        return jsonify({
            'metric_id': metric_id,
            'metric_name': metric_name,
            'anomalies': anomalies,
            'total_points': len(values),
            'anomaly_count': len(anomalies)
        })
        
    except Exception as e:
        logging.error(f"❌ Detection failed: {str(e)}")
        return jsonify({'error': str(e)}), 500

def detect_ensemble(timestamps, values):
    """Ensemble detection: Prophet + STL + Isolation Forest"""
    
    results = {
        'prophet': detect_prophet(timestamps, values),
        'stl': detect_stl(values),
        'isolation_forest': detect_isolation_forest(values)
    }
    
    # Weighted voting
    weights = {
        'prophet': 0.4,
        'stl': 0.3,
        'isolation_forest': 0.3
    }
    
    threshold = 0.6  # Lower threshold for better detection
    
    anomalies = []
    for i in range(len(timestamps)):
        score = 0.0
        detected_by = []
        
        for method, is_anomaly_array in results.items():
            if is_anomaly_array[i]:
                score += weights[method]
                detected_by.append(method)
        
        if score >= threshold:
            anomalies.append({
                'timestamp': int(timestamps[i]),
                'value': float(values[i]),
                'score': float(score),
                'methods': detected_by
            })
    
    return anomalies

def detect_prophet(timestamps, values):
    """Prophet-based anomaly detection"""
    try:
        df = pd.DataFrame({
            'ds': pd.to_datetime(timestamps, unit='s'),
            'y': values
        })
        
        model = Prophet(
            interval_width=0.95,
            daily_seasonality=False,
            weekly_seasonality=False,
            yearly_seasonality=False
        )
        
        model.fit(df)
        forecast = model.predict(df)
        
        # Anomaly if outside prediction interval
        is_anomaly = (df['y'] < forecast['yhat_lower']) | (df['y'] > forecast['yhat_upper'])
        return is_anomaly.values
        
    except Exception as e:
        logging.warning(f"Prophet failed: {e}, using fallback")
        return np.zeros(len(values), dtype=bool)

def detect_stl(values):
    """STL decomposition-based detection"""
    try:
        if len(values) < 24:
            return np.zeros(len(values), dtype=bool)
        
        stl = STL(values, seasonal=13, robust=True)
        result = stl.fit()
        residuals = result.resid
        
        # Anomaly if residual > 2.5 std deviations
        threshold = 2.5 * np.std(residuals)
        is_anomaly = np.abs(residuals) > threshold
        return is_anomaly
        
    except Exception as e:
        logging.warning(f"STL failed: {e}, using fallback")
        return np.zeros(len(values), dtype=bool)

def detect_isolation_forest(values):
    """Isolation Forest detection"""
    try:
        values_array = np.array(values).reshape(-1, 1)
        
        iso_forest = IsolationForest(
            contamination=0.1,
            random_state=42
        )
        
        predictions = iso_forest.fit_predict(values_array)
        is_anomaly = predictions == -1
        return is_anomaly
        
    except Exception as e:
        logging.warning(f"Isolation Forest failed: {e}, using fallback")
        return np.zeros(len(values), dtype=bool)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001, debug=True)
