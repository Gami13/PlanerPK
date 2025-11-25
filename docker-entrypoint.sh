#!/bin/sh
set -e

mkdir -p /app/data/current_timetable /app/data/downloads /app/data/logs

API_BASE_URL="${API_BASE_URL:-http://localhost:8080/api}"

cat > /app/static/env.js <<EOF
window.__RUNTIME_CONFIG__ = {
  API_BASE_URL: "${API_BASE_URL}"
};
EOF

exec google-calendar-timetable "$@"
