# GoogleCalendarTimetable

## Deployment model

PlanerPK is intended to run inside a container only. Supported deployments are:

- a Docker container built from the repository `Dockerfile`;
- a Nix container that provides the same runtime filesystem contract.

Running the Go binary, `go run`, `nix run`, or the NixOS systemd module directly on the host is not supported. The container must provide `/app/static` for frontend assets, a writable `/data` directory for persistent state, and the OAuth file at `/data/credentials.json`.

## Docker workflow

The repository ships with a multi-stage `Dockerfile`. It builds the SolidJS frontend, compiles the Go backend, and installs LibreOffice Calc for headless `.xls` to `.xlsx` conversion.

The container has two filesystem areas:

- `/app` contains the application and compiled static assets.
- `/data` contains all mutable runtime data and should be mounted as one persistent volume.

The runtime uses `/app` and `/data` instead of the XDG directory hierarchy. XDG is designed around per-user desktop applications and distributes files across host-specific locations, which is a poor fit for containerized deployments with explicit mounts. All mutable paths are therefore derived from `DATA_DIR`, keeping persistent state under `/data`.

The data directory has this layout:

```text
/data/
├── credentials.json
├── state/
│   ├── database.db
│   └── current_timetable_stats.json
├── logs/
├── current_timetable/
└── downloads/
```

The OAuth `credentials.json` file should be mounted read-only. It is never included in the image.

### Building-place cache upgrade

The resolver rebuilds classroom mappings from the current timetable legend and
only adds Google Maps links after confirming a Google Places result. Unresolved
classroom matches and rejected place candidates are logged while imports
continue without unsafe map links. See [Building place resolution](docs/building-place-resolution.md)
for matching, cache, and diagnostic behavior.

When upgrading from the legacy seeded building/classroom schema, remove
`/data/state/database.db` once before starting this version; the cache is
recreated from the imported legend.

### Build the image

```sh
docker build -t google-calendar-timetable .
```

### Run with Docker Compose

Example `docker-compose.yml` configuration:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    image: ghcr.io/gami13/google-calendar-timetable:latest
    container_name: google-calendar-timetable
    ports:
      - "8080:8080"
    environment:
      API_BASE_URL: "http://localhost:8080/api"
      LOG_LEVEL: "INFO"
      MINUTES_BETWEEN: "5"
      TIME_BETWEEN_SYNC_HOURS: "72"
      GOOGLE_PLACES_API_KEY: "${GOOGLE_PLACES_API_KEY}"
    volumes:
      - ./data:/data
      - ./credentials.json:/data/credentials.json:ro
    restart: unless-stopped
```

Start the application with:

```sh
docker compose up --build
```

The backend uses `/data` by default. `DATA_DIR` can be changed if a different mount point is required, but all derived data paths remain inside that directory. Static assets remain in `/app/static` and are not part of the persistent data volume.

### Runtime settings

The backend supports these environment variables:

- `DATA_DIR` — mutable data root; defaults to `/data`.
- `LOG_LEVEL` — `DEBUG`, `INFO`, `WARN`, or `ERROR`.
- `MINUTES_BETWEEN` — timetable update interval.
- `TIME_BETWEEN_SYNC_HOURS` — minimum calendar sync interval when the timetable has not changed.
- `TIMETABLE_URL` — source timetable page.
- `SHEET_NAME`, `HOUR_COLUMN`, `DAY_START`, `DAY_END`, and `LAST_COLUMN` — timetable spreadsheet layout settings.
- `GOOGLE_PLACES_API_KEY` — Google Places API key used to enrich building details.

The container listens on port `8080`. Change the host-facing port through the Compose mapping, for example `9000:8080`; the backend does not need a port environment variable. The frontend-only `API_BASE_URL` setting is substituted into the built frontend by the Docker entrypoint. It should be supplied by the deployment; the container does not invent a development URL. The static asset directory is fixed at `/app/static`.

Protobuf generation is run from the repository root with `buf generate --template buf.gen.yaml` and writes to `backend/gen` and `frontend/src/gen`.
