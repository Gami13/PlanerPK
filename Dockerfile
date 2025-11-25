# syntax=docker/dockerfile:1.7

###############################################
# Frontend build stage
###############################################
FROM node:22-bullseye AS frontend-builder
WORKDIR /app/frontend

# Install dependencies first for better layer caching
COPY frontend/pnpm-lock.yaml frontend/package.json ./
RUN corepack enable \
	&& pnpm install --frozen-lockfile

# Copy the remaining frontend sources and build the static bundle
COPY frontend/ ./
RUN pnpm build

###############################################
# Backend build stage
###############################################
FROM golang:1.24.4-bullseye AS backend-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY src ./src
RUN go build -o /app/bin/google-calendar-timetable ./src/main.go

###############################################
# Runtime stage
###############################################
FROM debian:bookworm-slim AS runtime

ENV APP_HOME=/app \
	HOME=/home/google
WORKDIR /app

# Install LibreOffice Calc (for headless conversions) and other runtime deps
RUN apt-get update \
	&& DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
	ca-certificates \
	libreoffice-calc \
	&& rm -rf /var/lib/apt/lists/*

# Create a unified data root and symlink legacy paths for compatibility
RUN mkdir -p data/current_timetable data/downloads data/logs \
	&& ln -s /app/data/current_timetable /app/current_timetable \
	&& ln -s /app/data/downloads /app/downloads \
	&& ln -s /app/data/logs /app/logs \
	&& ln -s /app/data/database.db /app/database.db

# Copy compiled artifacts and entry script
COPY --from=backend-builder /app/bin/google-calendar-timetable /usr/local/bin/google-calendar-timetable
COPY --from=frontend-builder /app/static ./static
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Use non-root user for better security
RUN useradd -r -u 10001 -g users -d /home/google -m google \
	&& mkdir -p /home/google/.cache/fontconfig /home/google/.cache/dconf /home/google/.config \
	&& chown -R google:users /app /home/google \
	&& chmod +x /usr/local/bin/docker-entrypoint.sh
USER google

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
