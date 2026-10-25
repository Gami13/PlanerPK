# Timetable diagnostics and ClassText scripts

Developer tools for inspecting matching, dates, event diffs, and classText samples.
Run commands from the `backend/` directory. These are not supported alternatives
to the application's Docker/Nix deployment.

## Snapshot subject index

Fetch the online snapshot and inspect its subject index (without importing or syncing):

```sh
go run ./cmd/indexsnapshot
```

The command takes no parameters and uses the published snapshot endpoint. Output is JSON; index totals go to stderr.

## Pipeline

### Offline event preview

`go run ./cmd/previewevents` reads one JSON object from stdin and prints the full
indented event diff, including before/after changes and recurring exceptions.
No network, OAuth, database access, or calendar writes are performed.

Input fields: `snapshot` (the decoded snapshot **state**, not its response envelope),
`timetable` (one GroupTimetable, with numeric weekday keys: Sunday=0),
`overrides` (optional, keyed by the existing class hash), and `currentEvents`
(optional array of captured Google events; defaults to an empty calendar).
Class and override fields use their Go/DB JSON representations. Building-place
enrichment is not loaded by this command.

```powershell
Get-Content -Raw preview-input.json | go run ./cmd/previewevents
```

### Matching and class text

`go run ./cmd/comparedates` displays each raw Excel weekday/time slot alongside
snapshot-derived meeting dates. It reads the current workbook and published JSON,
but does not apply user overrides or write calendars. There is no next-weekday
fallback or old-event generation.

`go run ./cmd/matchsnapshot` compares the current Excel workbook with the online snapshot.
Unique rows show only the group, subject/activity, and block ID; ambiguous and unmatched rows keep full details. Totals go to stderr.
It takes no parameters; `DATA_DIR` defaults to `../data`. Repeated classes across columns count separately.

```sh
go run ./cmd/collectclasstexts/
go run ./cmd/dedupclasstexts/
go run ./cmd/collapsespaces/
go run ./cmd/splitclasstexts/
go run ./cmd/splitbycount/
go run ./cmd/dumpclassrooms/
go run ./cmd/matchclassrooms/
```


