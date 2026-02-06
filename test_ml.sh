#!/bin/bash

curl -X POST http://localhost:5001/detect \
  -H "Content-Type: application/json" \
  -d '{
    "metric_id": 1,
    "metric_name": "test_metric",
    "timestamps": [1707251854, 1707251914, 1707251974, 1707252034, 1707252094, 1707252154, 1707252214, 1707252274, 1707252334, 1707252394, 1707252454, 1707252514],
    "values": [1.0, 1.1, 1.0, 1.2, 5.5, 1.1, 1.0, 1.1, 1.0, 1.2, 1.1, 1.0]
  }'
