#!/bin/sh
set -e

DATA_DIRECTORY="${DATA_DIRECTORY:-/data}"

# Ensure data subdirectories exist
mkdir -p "${DATA_DIRECTORY}/current_timetable" "${DATA_DIRECTORY}/downloads" "${DATA_DIRECTORY}/logs"

# Set ownership to google user
chown -R google:users "${DATA_DIRECTORY}"

# Create symlinks for backward compatibility
[ ! -e /app/current_timetable ] && ln -s "${DATA_DIRECTORY}/current_timetable" /app/current_timetable
[ ! -e /app/downloads ] && ln -s "${DATA_DIRECTORY}/downloads" /app/downloads
[ ! -e /app/logs ] && ln -s "${DATA_DIRECTORY}/logs" /app/logs
[ ! -e /app/database.db ] && ln -s "${DATA_DIRECTORY}/database.db" /app/database.db

echo "Using API_BASE_URL: ${API_BASE_URL}"

for js_file in /app/static/assets/*.js; do
  if [ -f "$js_file" ]; then
    envsubst '$API_BASE_URL' < "$js_file" > "${js_file}.tmp"
    mv "${js_file}.tmp" "$js_file"
  fi
done

# Switch to google user and run
exec gosu google google-calendar-timetable "$@"
