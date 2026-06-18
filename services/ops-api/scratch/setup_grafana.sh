#!/bin/bash
URL="http://127.0.0.1:3001"
AUTH="admin:admin"

echo "Adding Prometheus Data Source"
curl -s -u $AUTH -X POST $URL/api/datasources -H "Content-Type: application/json" \
-d '{"name":"Prometheus","type":"prometheus","url":"http://prometheus:9090","access":"proxy","isDefault":true}'

echo "Adding Jaeger Data Source"
curl -s -u $AUTH -X POST $URL/api/datasources -H "Content-Type: application/json" \
-d '{"name":"Jaeger","type":"jaeger","url":"http://jaeger:16686","access":"proxy"}'

cat << 'PY_EOF' > process_dash.py
import urllib.request
import json
import sys

dash_id = sys.argv[1]
url = f"https://grafana.com/api/dashboards/{dash_id}/revisions/latest/download"
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
with urllib.request.urlopen(req) as response:
    dash = json.loads(response.read().decode())

payload = {
    "dashboard": dash,
    "overwrite": True,
    "inputs": [{"name": "DS_PROMETHEUS", "type": "datasource", "pluginId": "prometheus", "value": "Prometheus"}]
}
dash['id'] = None
print(json.dumps(payload))
PY_EOF

echo "Importing Dashboard 1860 (Go Metrics)"
python3 process_dash.py 1860 > dash_1860.json
curl -s -u $AUTH -X POST $URL/api/dashboards/db -H "Content-Type: application/json" -d @dash_1860.json

echo "Importing Dashboard 14737 (Go HTTP RED Metrics)"
python3 process_dash.py 14737 > dash_14737.json
curl -s -u $AUTH -X POST $URL/api/dashboards/db -H "Content-Type: application/json" -d @dash_14737.json

