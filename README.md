# GoogleCalendarTimetable

## Docker Workflow

The repository now ships with a multi-stage `Dockerfile` that mirrors the native `build.ps1`/`build.sh` steps: it builds the SolidJS frontend with PNPM, compiles the Go backend, installs LibreOffice Calc for headless `.xls → .xlsx` conversions, and exposes the unified binary.

### Prerequisites

- Docker Engine or Docker Desktop 4.27+
- The OAuth `credentials.json` file present at the project root (it is bind-mounted at runtime, so it never becomes part of the image).
- Protobuf generation is run from repo root with `buf generate --template buf.gen.yaml` and writes to `backend/gen` and `frontend/src/gen`.

### Build the Image

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
      SERVER_PORT: "8080"
      API_BASE_URL: "http://localhost:8080/api" #base URL for the API that the frontend will communicate with. It should point to your backend server.
      SHEET_NAME: "2025_2026" #name of the sheet inside of the xls file
      LAST_ROW: "365" #last row of the sheet that contains data
      HOUR_COLUMN : 1 #column that contains the hour of the lesson
      DAY_START : "7.30" #time when the first lesson starts
      DAY_END : "21.00" #time when the last lesson ends
      DATABASE_SOURCE : "file:database.db?cache:shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite" #source of the database, can be a file or an in-memory database, time_format=sqlite is required
      LAST_COLUMN : "40" #last column of the sheet that contains data
      DATABASE_DRIVER : "sqlite" #driver for the database, can be sqlite or postgres
      MINUTES_BETWEEN : "5" #minutes between checking for new data
      GENERAL_RATE_LIMIT_PER_MINUT: "60" #general rate limit per minute for API requests
      CALENDAR_CREDENTIALS_PATH: /app/credentials.json
      TIMETABLE_DOWNLOAD_DESTINATION : "./downloads/"
      TIMETABLE_CURRENT_DESTINATION : "./current_timetable/" #destination folder for the current timetable, should be a folder inside the project
      TIMETABLE_URL : "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/" #URL of the timetable, should be a valid URL that points to the website with the timetable
      TIME_BETWEEN_SYNC_HOURS : "72" #time between syncing the calendar, in hours, if nothing changed
      TZ: "Europe/Warsaw" #timezone for the application
      
    volumes:
      - ./data:/app/data
      - ./credentials.json:/app/credentials.json:ro
    restart: unless-stopped
```


`docker-compose.yml` keeps everything under a single `./data` directory (downloads, current timetables, logs, database) in sync with the container, rewrites `static/env.js` at runtime so the frontend points to the correct API base, and exposes the HTTP API via the `SERVER_PORT` environment variable (defaults to `8080` if unset):

```sh
docker compose up --build
```

Key mounts defined in the Compose file:

- `./data:/app/data` → hosts the entire mutable state tree; the container creates the required subdirectories (`current_timetable`, `downloads`, `logs`, `database.db`) on startup if they are missing.
- `./credentials.json:/app/credentials.json:ro` → place your Google OAuth credentials file next to `docker-compose.yml`; it is mounted read-only inside the container.

During startup the entrypoint writes `/app/static/env.js` with the value of `API_BASE_URL` (defaulting to `http://localhost:8080/api`), so you can point the compiled frontend at different backend URLs without rebuilding the image.

Customize runtime settings by exporting environment variables (for example `SERVER_PORT`, `MINUTES_BETWEEN`, etc.) before running `docker compose` or by editing the `environment` block inside `docker-compose.yml`. Any variable supported by `src/global/global.go` can be overridden this way.


