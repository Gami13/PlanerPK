# Timetable importer handoff

Current status: 2026-10-25. See
[Hybrid timetable importer](snapshot-timetable-migration.md) for behavior and scope.
The earlier full-JSON replacement plan is superseded.

## Working agreement

- Keep changes small and reviewable, ideally under 100 changed lines including
  tests. Explain the step before editing; discuss larger splits. Run relevant
  checks, summarize, and stop for review. Commit only when requested.
- Reuse established names when new functions replace old ones. Avoid broad
  rewrites, excessive validation, generic frameworks, and unnecessary helpers.
- Preserve Excel group membership, supplemental-column merging, override keys,
  and optional building enrichment. Trust publisher fields; parse/check them where
  used rather than adding exhaustive state validation.
- Never delete user calendars, events, or production data as development cleanup.
  Normal sync deletions are limited to managed events in the saved target calendar.
- The user accepted a fresh database for calendar_id; no schema migration or
  automatic database deletion was requested. Do not recreate accounts on login.
- Runtime deployment is Docker/Nix. Diagnostic Go commands are developer tools,
  not an alternative supported deployment. See backend/cmd/README.md.

## Current code

- backend/extractor contains all workbook parsing, including heuristic parsing
  and cadence-based layout detection formerly in backend/legacy. There are no
  remaining legacy-package imports. XLS conversion/LibreOffice are removed.
- Matching belongs in backend/snapshot. Only unique matches attach
  Class.SnapshotBlockID. Import warnings leave unresolved IDs empty; event
  construction fails before calendar writes rather than guessing dates.
- Cache has independent atomic timetable and snapshot fields, as explicitly
  chosen by the user. Do not introduce a paired generation wrapper or cache mutex.
  Load one snapshot reference per operation; never refetch/rematch per class.
- Class.ToEvent requires a ClassSchedule value. buildClassEvents and
  buildTimetableEvents return []*calendar.Event. Nil resolved schedules mean no
  meetings; no next-weekday fallback remains.
- Both live callers use the new dedicated-calendar pipeline. User settings and
  overrides are reloaded under a per-user sync lock within one application process.
- Overrides with a nonzero source_column import a class from another group in the
  same academic year. Compose them with UserTimetable before building events; never
  mutate the cached timetable.
- Google identities use overrideId (existing class hash, with _ghost for ghosts).
  Snapshot block IDs are not assumed stable across publications. Actual event
  contents are compared; contentHash is not stored and old values are ignored.
- Calendar creation uses the saved ID or creates a new PlanerPK calendar. There is
  no name search or automatic recovery for a missing/inaccessible saved calendar.
  OAuth requests only `calendar.app.created` (plus `email`/`profile`): create the
  secondary calendar and manage its events, never read the user's other calendars.
- List all pages of managed series and exceptions, not expanded instances. Block
  the entire write batch if it changes/deletes a series with recurring exceptions;
  unrelated series can sync. Do not silently discard manual meeting changes.
- Writes fail fast and last_synced advances only after all writes succeed.
  There is no rollback; retries list current state and calculate a fresh diff.
- Rebuild the combined cache each poll, downloading Excel only when needed.
  Compare the XLSX byte hash and publisher's snapshot.sha256 to the in-memory
  cached hashes; no deep object comparisons or stats-file hash persistence.
  Skip scheduled Google sync when unchanged; missing hashes trigger sync.
  Failed batches remain pending for retry; startup checks existing users once.
  Group/override changes trigger that user's sync directly, without proposal caches.
  Write logs include actual created/updated/deleted/skipped counts and
  batch success; exception guards identify the blocked parent series.
- The process timezone is supplied via Docker Compose by the user; recurring
  Google event start/end also specify Europe/Warsaw. No timezone rewrite requested.

## Matching cautions

- Excel owns regular/mixed-language/elective membership. JSON metadata cannot
  reproduce all supplemental-column behavior. Do not identify languages by the
  substring Język or invent fuzzy SIR/SIRR mappings.
- Matching uses subject/activity/weekday/start, teacher, understood audience
  metadata, and first-line A/B or explicit-date notes. A teacher-qualified subject
  fallback runs only when ordinary matching finds no candidates. Never choose an
  ambiguous candidate arbitrarily. End time and room are not matching filters.
- Keep scheduling notes as AdditionalInfo; do not reinterpret arbitrary text as
  date rules. Treat blocks independently, without inventing complementary A/B
  attendance. Fixture parity is 0=A, 1=B; use the supplied parity anchor.
- calendarExport is an offline oracle, never a production scheduling input.
  Fixture counts and historical live matching totals are observations, not
  permanent acceptance criteria. Compare dates by block and UTC start/end.

## Verification and remaining scope

The user manually verified the live cutover. Latest full backend checks passed:
go test ./... and git diff --check. Tests use mock Google HTTP responses and
in-memory SQLite; no real calendars were touched by development tools.
Race testing was attempted but blocked by the local Windows CGO toolchain.
Docker/Nix builds were not run in this cleanup.

Deferred features remain deferred; see the behavior reference. Documentation and
small test/setup cleanup are appropriate follow-ups, not an invitation to expand
calendar behavior or rebuild persistence infrastructure.
