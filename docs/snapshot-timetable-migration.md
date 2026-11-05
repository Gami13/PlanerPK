# Hybrid timetable importer — current behavior

Updated 2026-10-25 after the live cutover and legacy cleanup. This replaces the
earlier migration checklist and full-JSON replacement proposal.

## Sources and ownership

- Excel supplies class details and group membership:
  https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/download.php
- JSON supplies scheduling metadata:
  https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php
- Production reads state.blocks and state.academicCalendar. calendarExport is
  only an offline test oracle.

The existing Excel group picker is retained. Columns marked językowe/obieralne
are merged into regular groups of the same category and excluded from the picker.
Users hide unwanted alternatives through existing overrides. JSON cohorts are
not used to reconstruct membership. Raw Excel scheduling notes remain available
as AdditionalInfo rather than being interpreted as general date rules.

All workbook code lives in backend/extractor, including the older heuristic
format. Downloads are XLSX; no XLS conversion or LibreOffice is needed.

## Import, matching, and cache

The importer loads JSON, parses Excel, and attaches SnapshotBlockID only for
unique matches. Unmatched/ambiguous classes warn during import, retain an empty
ID, and fail event construction before any calendar writes. There is no guessed
next-weekday schedule.

Matching uses subject, activity, weekday, start time, and understood teacher,
audience, A/B, and explicit-date metadata. A teacher-qualified cross-subject
fallback is used only when ordinary matching returns zero candidates. End time
and room are not matching filters; unknown metadata must not force an arbitrary
choice. Block IDs link records within an import, not across publications.

Cache retains independent atomic timetable and *snapshot.State fields. A complete
replacement is built before publication; failed builds retain the previous cache.
Each operation loads the snapshot once and passes that reference downstream.
Sync neither refetches JSON nor rematches classes. Startup requires usable online
inputs; there is no persisted JSON fallback. An upstream HTTP failure can prevent
startup before Google Calendar is contacted.

Each poll rebuilds from freshly fetched JSON and the current workbook. Excel is
downloaded only when its update check reports a change (a missing Last-Modified
header also triggers a download). Compare the current XLSX byte hash and the
publisher-provided snapshot.sha256 against the hashes in the successfully built
cache. Identical XLSX re-downloads do not trigger Google checks, even when headers
are absent or changed. No per-block/deep comparisons or stats-file hash persistence
are needed. This trusts the publisher to update its snapshot fingerprint; a missing
snapshot hash conservatively triggers sync. If neither hash changes, skip scheduled
Google requests entirely. Workbook hashes advance only after a successful rebuild,
so failed builds do not acknowledge new content. Failed batches
remain pending for retry, and startup checks existing users once. Group/override
changes trigger that user's sync directly. There is no per-user proposal cache or
sync-hours throttle. Failed rebuilds retain the cache and skip that poll's sync.

## Scheduling currently supported

- Weekly meetings and alternating A/B weeks, using teachingWeekParity and
  parityAnchor, not ISO week parity or hardcoded semester dates.
- First/second semester halves using period and halfBoundary.
- Term bounds, full-day closures, and overlapping partial-day closures.
- Concrete-date meetings (konkretna_data) as single events.
- Weekly/A-B/half-semester meetings as recurring series with bounded RRULE and
  EXDATE for closures. An entirely empty schedule emits nothing.
- Parked blocks with missing day/start emit nothing; dates are not guessed.
- jednorazowo and pierwszy_tydzien_pazdziernika are intentionally omitted by
  user choice. This omission does not apply to konkretna_data.

Local date generation uses the process timezone, configured as Europe/Warsaw by
the deployment's Docker Compose environment. Recurring Google start/end fields
also explicitly specify Europe/Warsaw to preserve local time across DST.

ClassSchedule is separate from Class and Google's Event model. ToEvent requires
a schedule value. Hidden, shifted, and transparent ghost overrides apply to the
whole class schedule. A ghost stays at the original slot; its real event can
move. Event collections consistently use []*calendar.Event.

## Calendar ownership and synchronization

Each user has a persisted users.calendar_id. A saved ID is authoritative; without
one, create a PlanerPK secondary calendar and save its ID before event sync.
There is no search for an existing calendar by name and no writes to primary.
Only `calendar.app.created` (plus `email`/`profile`) is requested, so the app can
create its own secondary calendar and manage events on it, but cannot read the
user's other calendars. An inaccessible saved calendar causes an error, not
automatic recreation.

Both scheduled refresh and group selection use this flow:

1. Serialize sync per user within the application process; reload current user
   settings and overrides after obtaining the lock. Override-read errors abort.
2. Build all proposed events from one loaded snapshot; build failures abort.
3. Ensure/persist the target calendar ID and list all pages of managed series and
   exceptions, without expanding instances or applying a rolling today cutoff.
4. Match by private overrideId: existing class hash for real events, plus _ghost
   for ghosts. Existing database override keys are unchanged. Duplicate or missing
   identities fail the diff instead of silently overwriting entries.
5. Compare actual importer-owned contents: dates, times, recurrence, timezone,
   displayed details, transparency, and color. Equivalent timestamp offsets and
   Google's default opaque transparency do not trigger changes. No contentHash
   is stored; old stored hashes are ignored.
6. Keep modified/cancelled recurring exceptions separate. Refuse the entire write
   batch if it would update/delete their parent series. Unrelated series may sync.
7. Insert additions, patch changes, then delete removed managed events. Patch only
   importer-owned fields; preserve unrelated fields such as user reminders.
8. Stop at the first write failure. Do not advance last_synced until all writes
   succeed. There is no rollback; retries re-list and compare current state.

Write summaries log successful created/updated/deleted counts, unchanged/skipped
counts, and whether the write batch succeeded. Failures report partial counts,
not planned totals. Exception guards log the blocked parent series ID.

Only events carrying the importer marker, or exceptions belonging to managed
series, are included. Deleted standalone events/series are not live diff targets.
Per-user locking is process-local, not coordination between multiple replicas.

## Offline review and verification

backend/cmd/previewevents accepts captured timetable, snapshot, overrides, and
existing Google events on stdin and prints a full JSON diff. It performs no
network requests, OAuth, database operations, or calendar writes. See
[Command documentation](../backend/cmd/README.md) for input details and usage.

comparedates shows raw Excel weekday/time slots alongside snapshot occurrences;
it no longer constructs the old next-weekday events. Matching/index/date commands
remain diagnostic tools. Full app deployment is supported through Docker.

Backend tests cover builders, schedules, overrides, content comparison,
pagination, exceptions, mocked writes, calendar-ID persistence, concurrent sync,
and retry after partial writes. The user manually verified the live cutover.
Captured fixture counts are evidence, not requirements or a claim that all
publisher exceptions are implemented.

## Deferred features and limits

Do not silently expand scope to implement these during cleanup:

- Replacement days/RDATE, regular stationary start, per-block allowed/excluded
  dates, and their precedence. DTO fields exist but are not currently applied.
- Automatic reconciliation of individually modified/cancelled recurring meetings.
- Persisted snapshot fallback for offline startup.
- Cross-publication matching based on supposedly stable publisher block IDs.
- Calendar rediscovery, old-event migration, and multi-replica sync coordination.

## Deployment and data safety

The user accepted a fresh database for the calendar_id schema change; no database
migration is included. CREATE TABLE IF NOT EXISTS does not upgrade an old users
table. Provision a fresh database explicitly when upgrading from that schema.
Normal logins upsert users and preserve their saved calendar IDs.

No automatic database reset or calendar/event deletion is used for development
cleanup. Existing calendars are never deleted. Normal sync can remove obsolete
importer-managed events from its target calendar only. Building enrichment remains
optional; existing location and override behavior is preserved.
