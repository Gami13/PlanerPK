# GoogleCalendarTimetable

## Deployment model

PlanerPK is intended for managed deployment only. Supported deployments are:

- a Docker container built from the repository `Dockerfile`;
- the NixOS module, which provides the same runtime filesystem contract.

Running the Go binary, `go run`, or `nix run` directly on the host is not supported. The runtime serves frontend assets from its internal `static` directory, stores mutable data under `DATA_DIR` (default `/data`), and reads OAuth credentials from `DATA_DIR/credentials.json`.

## Docker workflow

The repository ships with a multi-stage `Dockerfile`. It builds the SolidJS frontend and compiles the Go backend. Timetables are downloaded directly as `.xlsx` files.

The container has two filesystem areas:

- `static` contains the internal compiled frontend assets.
- `/data` contains all mutable runtime data and should be mounted as one persistent volume.

The runtime does not use the XDG directory hierarchy. All mutable paths are derived from `DATA_DIR`, while the static directory remains internal to the deployment.

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

The OAuth `credentials.json` file should be mounted read-only. It is never included in the image. In Google Cloud, the OAuth consent screen requests `calendar.app.created`, `userinfo.email`, and `userinfo.profile`, and its privacy policy URL points at the public `/privacy-policy` route (for example `https://planpk.gami.ovh/privacy-policy`).

### Timetable calendar behavior

Excel provides class details and group membership; the university JSON snapshot
provides dated schedules. Events sync to a dedicated **PlanerPK** calendar whose
ID is saved per user, not to the primary calendar. Weekly classes use recurring
series; concrete-date classes remain single events. Existing hide, time-shift,
and transparent ghost overrides are preserved.

Startup needs usable upstream Excel and JSON inputs. There is no offline JSON
fallback. Failed syncs are not marked successful; changes to series with manual
recurring exceptions stop for review rather than silently discarding exceptions.

When upgrading from a schema without `users.calendar_id`, provision a fresh
database explicitly. No migration or automatic reset is included; keep OAuth
credentials and other deployment files intact. Normal logins preserve users and
their saved calendar IDs. No existing calendars are deleted as upgrade cleanup.

See [Importer behavior and deferred features](docs/snapshot-timetable-migration.md)
and [diagnostic commands](backend/cmd/README.md), including the offline diff preview.

### Building-place cache upgrade

Building-place resolution is controlled by `BUILDING_LOCATIONS_ENABLED` and is
disabled by default. When enabled, the resolver rebuilds classroom mappings
from the current timetable legend and only adds Google Maps links after
confirming a Google Places result. Unresolved classroom matches and rejected
place candidates are logged while imports continue without unsafe map links.
When disabled, raw classroom/location text is still included in calendar events,
but building mappings, place-cache updates, Google Places lookups, and map links
are skipped. See [Building place resolution](docs/building-place-resolution.md)
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
      TZ: "Europe/Warsaw"
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

The backend uses `DATA_DIR=/data` by default. Every mutable runtime path is derived from it; `static` remains internal to the deployment.

### Runtime settings

The backend supports these environment variables:

- `DATA_DIR` — mutable data root; defaults to `/data`.
- `LOG_LEVEL` — `DEBUG`, `INFO`, `WARN`, or `ERROR`.
- `MINUTES_BETWEEN` — combined timetable refresh interval. Unchanged cache contents skip scheduled Google sync; group/override changes sync directly.
- `TZ` — container process timezone; use `Europe/Warsaw` for local class dates and DST.
- `TIMETABLE_URL` — direct timetable download URL.
- `GOOGLE_PLACES_API_KEY` — Google Places API key used to enrich building details.
- `BUILDING_LOCATIONS_ENABLED` — feature flag for classroom-to-building matching, place-cache updates, Google Places lookups, and map links; defaults to `false`.
- `DISCORD_WEBHOOK_URL` — optional Discord webhook used for error notifications.

The active worksheet is selected automatically using the `YYYY_YYYY` academic-year naming convention, with the academic-year transition on September 15. On each import, the timetable hour column, slot duration, and day boundaries are inferred from consistent time sequences in the worksheet; they are no longer configured through environment variables.

The container listens on port `8080`. Change the host-facing port through the Compose mapping, for example `9000:8080`; the backend does not need a port environment variable. The frontend-only `API_BASE_URL` setting is substituted into the built frontend by the Docker entrypoint. It should be supplied by the deployment, for example `https://planpk.gami.ovh/api`.

Protobuf generation is run from the repository root with `buf generate --template buf.gen.yaml` and writes to `backend/gen` and `frontend/src/gen`.
