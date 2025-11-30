#!/bin/sh
set -e

# Ensure data subdirectories exist (volume may be empty on first run)
mkdir -p /app/data/current_timetable /app/data/downloads /app/data/logs

# Set ownership to google user
chown -R google:users /app/data

# Create symlinks for backward compatibility (only if they don't exist)
[ ! -e /app/current_timetable ] && ln -s /app/data/current_timetable /app/current_timetable
[ ! -e /app/downloads ] && ln -s /app/data/downloads /app/downloads
[ ! -e /app/logs ] && ln -s /app/data/logs /app/logs
[ ! -e /app/database.db ] && ln -s /app/data/database.db /app/database.db
echo "Using API_BASE_URL: ${API_BASE_URL}"

#TODO: replace this with envsubst 

cat > /app/static/env.js <<EOF
window.__RUNTIME_CONFIG__ = {
  API_BASE_URL: "${API_BASE_URL}"
};
EOF

# Switch to google user and run the application
exec gosu google google-calendar-timetable "$@"
