#!/usr/bin/env bash
if [ -f .env ]; then
  set -o allexport
  source .env
  set +o allexport
  echo "Loaded environment variables from .env"
fi

go -C ./backend run .
