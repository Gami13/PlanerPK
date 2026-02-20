#!/bin/sh
set -e

# Ensure data subdirectories exist
mkdir -p /app/data/current_timetable /app/data/downloads /app/data/logs

# Set ownership to google user
chown -R google:users /app/data

# Create symlinks for backward compatibility
[ ! -e /app/current_timetable ] && ln -s /app/data/current_timetable /app/current_timetable
[ ! -e /app/downloads ] && ln -s /app/data/downloads /app/downloads
[ ! -e /app/logs ] && ln -s /app/data/logs /app/logs
[ ! -e /app/database.db ] && ln -s /app/data/database.db /app/database.db

echo "Using API_BASE_URL: ${API_BASE_URL}"

for js_file in /app/static/assets/*.js; do
  if [ -f "$js_file" ]; then
    envsubst '$API_BASE_URL' < "$js_file" > "${js_file}.tmp"
    mv "${js_file}.tmp" "$js_file"
  fi
done

# Switch to google user and run
exec gosu google google-calendar-timetable "$@"
