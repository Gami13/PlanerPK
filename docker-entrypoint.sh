#!/bin/sh
set -e

# Set ownership to google user
chown -R google:users /planer-pk || true

echo "Using API_BASE_URL: ${API_BASE_URL}"

for js_file in /planer-pk/static/assets/*.js; do
  if [ -f "$js_file" ]; then
    envsubst '$API_BASE_URL' < "$js_file" > "${js_file}.tmp"
    mv "${js_file}.tmp" "$js_file"
  fi
done

# Switch to google user and run
exec gosu google planer-pk "$@"
