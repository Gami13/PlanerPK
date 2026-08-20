#!/bin/sh
set -e

# Set ownership for mutable data without changing the read-only credentials file.
find "$DATA_DIR" -path "$DATA_DIR/credentials.json" -prune -o -exec chown google:users {} +

echo "Using API_BASE_URL: ${API_BASE_URL}"

for js_file in /app/static/assets/*.js; do
  if [ -f "$js_file" ]; then
    envsubst '$API_BASE_URL' < "$js_file" > "${js_file}.tmp"
    mv "${js_file}.tmp" "$js_file"
  fi
done

# Switch to google user and run
exec gosu google planer-pk "$@"
