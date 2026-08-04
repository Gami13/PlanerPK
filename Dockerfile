# syntax=docker/dockerfile:1.7

# --- Frontend ---
FROM node:22-alpine AS frontend-builder
WORKDIR /planer-pk/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
	corepack enable && pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# --- Backend ---
FROM golang:1.24-alpine AS backend-builder
WORKDIR /planer-pk/backend
ARG TARGETOS
ARG TARGETARCH
COPY backend/go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download
COPY backend/ ./
RUN --mount=type=cache,target=/root/.cache/go-build \
	GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
	go build -ldflags="-s -w" -o /planer-pk/bin/main ./main.go

# --- Runtime ---
FROM debian:bookworm-slim AS runtime
ENV APP_HOME=/planer-pk HOME=/home/google
WORKDIR /planer-pk

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
	rm -f /etc/apt/apt.conf.d/docker-clean && \
	apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
	ca-certificates libreoffice-calc gosu gettext-base && \
	useradd -r -u 1000 -g users -d /home/google -m google && \
	mkdir -p /home/google/.config && \
	chown -R google:users /home/google

COPY --link --from=backend-builder /planer-pk/bin/main /usr/local/bin/planer-pk
COPY --link --from=frontend-builder /planer-pk/static ./static
COPY --chmod=755 docker-entrypoint.sh /usr/local/bin/

# Remove Windows line endings causing the "no such file" error
RUN sed -i 's/\r$//' /usr/local/bin/docker-entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
