# GoogleCalendarTimetable

## Docker Workflow

The repository now ships with a multi-stage `Dockerfile` that mirrors the native `build.ps1`/`build.sh` steps: it builds the SolidJS frontend with PNPM, compiles the Go backend, installs LibreOffice Calc for headless `.xls → .xlsx` conversions, and exposes the unified binary.

### Prerequisites

- Docker Engine or Docker Desktop 4.27+
- The OAuth `credentials.json` file present at the project root (it is bind-mounted at runtime, so it never becomes part of the image).

### Build the Image

```sh
docker build -t google-calendar-timetable .
```

### Run with Docker Compose

`docker-compose.yml` keeps everything under a single `./data` directory (downloads, current timetables, logs, database) in sync with the container, rewrites `static/env.js` at runtime so the frontend points to the correct API base, and exposes the HTTP API via the `SERVER_PORT` environment variable (defaults to `8080` if unset):

```sh
docker compose up --build
```

Key mounts defined in the Compose file:

- `./data:/app/data` → hosts the entire mutable state tree; the container creates the required subdirectories (`current_timetable`, `downloads`, `logs`, `database.db`) on startup if they are missing.
- `./credentials.json:/app/credentials.json:ro` → place your Google OAuth credentials file next to `docker-compose.yml`; it is mounted read-only inside the container.

During startup the entrypoint writes `/app/static/env.js` with the value of `API_BASE_URL` (defaulting to `http://localhost:8080/api`), so you can point the compiled frontend at different backend URLs without rebuilding the image.

Customize runtime settings by exporting environment variables (for example `SERVER_PORT`, `MINUTES_BETWEEN`, etc.) before running `docker compose` or by editing the `environment` block inside `docker-compose.yml`. Any variable supported by `src/global/global.go` can be overridden this way.