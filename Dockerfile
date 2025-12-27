# syntax=docker/dockerfile:1.7

# --- Frontend ---
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
	corepack enable && pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# --- Backend ---
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download
COPY src ./src
RUN --mount=type=cache,target=/root/.cache/go-build \
	GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
	go build -ldflags="-s -w" -o /app/bin/main ./src/main.go

# --- Runtime ---
FROM debian:bookworm-slim AS runtime
ENV APP_HOME=/app HOME=/home/google
WORKDIR /app

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
	rm -f /etc/apt/apt.conf.d/docker-clean && \
	apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
	ca-certificates libreoffice-calc gosu gettext-base && \
	useradd -r -u 1000 -g users -d /home/google -m google && \
	mkdir -p data/current_timetable data/downloads data/logs /home/google/.config && \
	chown -R google:users /app/data /home/google

COPY --link --from=backend-builder /app/bin/main /usr/local/bin/google-calendar-timetable
COPY --link --from=frontend-builder /app/static ./static
COPY --chmod=755 docker-entrypoint.sh /usr/local/bin/

# FIX: Remove Windows line endings causing the "no such file" error
RUN sed -i 's/\r$//' /usr/local/bin/docker-entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]