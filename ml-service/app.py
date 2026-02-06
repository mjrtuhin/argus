from flask import Flask, request, jsonify
import numpy as np
from sklearn.ensemble import IsolationForest
import logging

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)

@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'healthy', 'service': 'argus-ml'})

@app.route('/detect', methods=['POST'])
def detect_anomalies():
    try:
        data = request.json
        
        # Extract data
        metric_id = data.get('metric_id')
        metric_name = data.get('metric_name')
        timestamps = data.get('timestamps', [])
        values = data.get('values', [])
        
        if len(values) < 10:
            return jsonify({
                'error': 'Need at least 10 data points',
                'anomalies': []
            }), 400
        
        # Run Isolation Forest
        anomalies = detect_isolation_forest(timestamps, values)
        
        logging.info(f"Detected {len(anomalies)} anomalies for metric {metric_name}")
        
        return jsonify({
            'metric_id': metric_id,
            'metric_name': metric_name,
            'anomalies': anomalies,
            'total_points': len(values),
            'anomaly_count': len(anomalies)
        })
        
    except Exception as e:
        logging.error(f"Detection failed: {str(e)}")
        return jsonify({'error': str(e)}), 500

def detect_isolation_forest(timestamps, values):
    """Detect anomalies using Isolation Forest"""
    
    # Reshape values for sklearn
    values_array = np.array(values).reshape(-1, 1)
    
    # Train Isolation Forest
    iso_forest = IsolationForest(
        contamination=0.1,  # Expect 10% anomalies
        random_state=42
    )
    
    predictions = iso_forest.fit_predict(values_array)
    scores = iso_forest.score_samples(values_array)
    
    # Find anomalies (prediction = -1)
    anomalies = []
    for i, (pred, score) in enumerate(zip(predictions, scores)):
        if pred == -1:
            anomalies.append({
                'timestamp': int(timestamps[i]),
                'value': float(values[i]),
                'score': abs(float(score)),  # Higher = more anomalous
                'methods': ['isolation_forest']
            })
    
    return anomalies

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001, debug=True)
