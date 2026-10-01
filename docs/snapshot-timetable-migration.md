# Snapshot timetable migration — implementation handoff

## Goal and boundaries

Replace the Excel importer with a JSON importer for
`https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php`.
**Production derives groups, classes, and dated occurrences from `state` only.**
Use `calendarExport.events` from a captured response **only as a test oracle**
for the date rules. Do not build production calendars from those exported events.

This is a **fresh-start database deployment**: new group IDs, no conversion of
saved Excel columns, users or overrides. The application should remain functional on a
fresh `/data` without any Excel file. Move the old extractor to legacy, preferably a sub-module.

Observed response (2026-10-01): JSON object with `state`, `timetableNotices`,
`validation`, and `calendarExport`. `state.schemaVersion` is `1`; 369 blocks,
25 rooms, and an `academicCalendar` for winter 2026/27. The separate export has
4,058 dated events covering 363 block IDs; six unscheduled blocks have no
events. These are fixture observations, **not** required counts for future
snapshots. Response is ~1.5 MB; it advertises `Cache-Control: no-store` and
has no observed `Last-Modified` or `ETag`. Do not rely on `HEAD` or file size.

## Proposed Go wire types

Put source DTOs in a new package, e.g. `backend/snapshot`; keep them separate
from domain `types.Class` and Google Calendar event types. These are suggested
shapes rather than an instruction to mirror every JSON property. Use JSON tags,
`*T` for nullable/absent scalars, and explicit date/time parsing during
validation. Unknown additive fields should not break decoding. Never infer a
scheduled date from an absent field.

```go
type Response struct {
    State          State          `json:"state"`
    CalendarExport *CalendarExport `json:"calendarExport"` // tests only; omit in production DTO if desired
}
type State struct {
    SchemaVersion    int              `json:"schemaVersion"`
    PublicationMode  string           `json:"publicationMode"`
    Blocks           []Block          `json:"blocks"`
    Rooms            []Room           `json:"rooms"`
    AcademicCalendar AcademicCalendar `json:"academicCalendar"`
    Meta             StateMeta        `json:"meta"`
    Snapshot         SnapshotMeta     `json:"snapshot"`
    // Other source fields (constraints, wishes, settings, sourceSummary,
    // legacy) can be ignored until there is a demonstrated need for them.
}
type Block struct {
    ID                 string           `json:"id"`
    Subject            string           `json:"subject"`
    Activity           string           `json:"activity"` // e.g. W, C, L, P, W/C, W+L, S, WF
    Teacher            string           `json:"teacher"`
    TeacherDisplay     string           `json:"teacherDisplay"`
    Cohort             string           `json:"cohort"`
    AdditionalCohorts  []string         `json:"additionalCohorts"` // absent or null is allowed
    PlanType           string           `json:"planType"`
    Day                *string          `json:"day"` // MON..FRI; null for parked blocks
    Start              *int             `json:"start"` // minutes since local midnight
    Duration           int              `json:"duration"` // minutes
    Room               *string          `json:"room"`
    Campus             *string          `json:"campus"`
    Modality           string           `json:"modality"`
    Hours              *int             `json:"hours"`
    Frequency          string           `json:"frequency"`
    Period             string           `json:"period"`
    Date               *string          `json:"date"` // YYYY-MM-DD, for a dated block
    TeachingWeekParity *int             `json:"teachingWeekParity"` // observed 0, 1 or null
    GroupNo            *int             `json:"groupNo"`
    StudentGrouping    *StudentGrouping `json:"studentGrouping"`
    AllowedDates       []string         `json:"allowedDates"`
    ExcludedDates      []string         `json:"excludedDates"`
    AcademicGroupLabel string           `json:"academicGroupLabel"`
    SharedSpecialties  []string         `json:"sharedSpecialties"`
    ParkingReason      string           `json:"parkingReason"`
    // Add other display/selection fields only when their semantics are tested.
}
type StudentGrouping struct {
    Kind              string   `json:"kind"` // observed mixed_language, masters_academic
    Pool              string   `json:"pool"`
    Group             *int     `json:"group"`
    TotalGroups       *int     `json:"totalGroups"`
    Version           string   `json:"version"`
    GroupIDs          []string `json:"groupIds"`
    LabGroupIDs       []string `json:"labGroupIds"`
    Specialties       []string `json:"specialties"`
    LegacyExcessGroup bool     `json:"legacyExcessGroup"`
}
type AcademicCalendar struct {
    Year                   string            `json:"year"`
    Season                 string            `json:"season"`
    Start                  string            `json:"start"` // YYYY-MM-DD
    End                    string            `json:"end"`
    Closures               []DateRange       `json:"closures"`
    PartialClosures        []PartialClosure  `json:"partialClosures"`
    ReplacementDays        map[string]string `json:"replacementDays"` // actual date -> replaced date
    ParityAnchor           string            `json:"parityAnchor"`
    FirstWeek              string            `json:"firstWeek"` // observed "A — nieparzysty"
    HalfBoundary           string            `json:"halfBoundary"`
    RegularStationaryStart string            `json:"regularStationaryStart"`
}
type DateRange struct {
    From string `json:"from"`
    To   string `json:"to"`
}
type PartialClosure struct {
    Date  string `json:"date"`
    Start int    `json:"start"` // minutes from midnight
    End   int    `json:"end"`
}
type Room struct {
    Code     string `json:"code"`
    Name     string `json:"name"`
    Campus   string `json:"campus"`
    RoomType string `json:"room_type"`
    Capacity int    `json:"capacity"`
    Active   int    `json:"active"` // observed 0/1, not JSON booleans
}
type StateMeta struct {
    Term          string `json:"term"`
    SourceRefresh string `json:"sourceRefresh"`
}
type SnapshotMeta struct {
    SavedAt string `json:"savedAt"`
    SHA256  string `json:"sha256"`
}
type CalendarExport struct {
    Version  int           `json:"version"`
    Timezone string        `json:"timezone"`
    Events   []ExportEvent `json:"events"`
}
type ExportEvent struct {
    BlockID string `json:"blockId"`
    UID     string `json:"uid"`
    Start   string `json:"start"` // UTC YYYYMMDDTHHMMSSZ
    End     string `json:"end"`
}
```

`Block` fields vary: `additionalCohorts` may be missing/null/array;
`studentGrouping` occurs on only some blocks; `date`, `day`, `start`, `room`,
and `teachingWeekParity` can be null. `allowedDates` and `excludedDates`
matter to occurrence rules. A few blocks with null day/start are deliberately
parked, not corrupt. do not assume the Excel legend still
exists. `state.snapshot.sha256` is publisher metadata: do **not** assume it is
the SHA-256 of the HTTP body without verifying its specification. Calculate
our own change fingerprint from the validated source data.

## Ordered function-level work

Implement in this order, with focused tests and review after each section.
Names/signatures are proposals; explicit dependencies (`http.Client`, clock,
paths) make functions testable without hitting the live endpoint.

### 1. Acquire and validate a source state

1. `fetchSnapshot(ctx, client, url) ([]byte, error)`: bounded GET, timeout,
   check HTTP status and content type, limit response size before decoding.
   Avoid changing active state on network errors.
2. `decodeSnapshot(data) (Response, error)`: decode the envelope; tests should
   ensure root `state` is present (a missing object must not decode to an
   apparently valid zero-value state). Capture a **sanitized, committed fixture**
   and oracle export for offline tests; tests must not depend on the live URL.
3. `validateState(state) error`: accept only supported `schemaVersion` and
   publication mode; validate IDs unique/nonempty, required calendar dates and
   ordering, supported frequencies/periods/parity, and scheduled block times.
   Parse dates strictly and use `Europe/Warsaw`. Validate one-off dates and
   allowed/excluded dates. A parked block may lack a day/start, but must never
   acquire a guessed occurrence. Reject unknown scheduling rules rather than
   silently importing the wrong calendar. Validate that a timetable expected
   to have classes is not unexpectedly empty.
4. `loadLastGoodSnapshot(path)` and `saveLastGoodSnapshot(path, state)`: persist
   only a successfully validated **and built** state, via a temporary file and
   atomic rename. On fetch/parse/build failures retain the in-memory cache;
   define startup behavior if there is no valid persisted state (fail startup
   rather than publish an empty timetable). State file is under `DATA_DIR`.

**Review:** wire types, validation error cases, fixture contents.

### 2. Determine selectable groups and membership

5. `buildSelectableGroups(state) ([]Group, error)`: derive a stable string ID,
   display label, and category for each user-selectable group. Never use a
   block-array position, `groupNo` alone, or an old Excel column number as an
   ID. Do not expose every raw `cohort` as a choice without checking parent
   cohorts, language pools, Erasmus, electives, and master's specialties.
6. `blockBelongsToGroup(block, group) bool`: the **single** membership policy
   shared by UI class lists and calendar sync. Test exact cohorts, parent
   cohort lectures, `additionalCohorts`, mixed-language pools,
   `masters_academic.groupIds`/`labGroupIds`, and specialty sharing. Names such
   as `CY`, `DS`, `SIR` and `SIRR` require evidence-based mapping, not fuzzy
   string containment. Establish whether language/elective blocks are included
   for all relevant students or offered as separately selectable groups; show
   concrete fixture examples to the reviewer before locking this in.
7. `blocksForGroup(state, groupID) ([]Block, error)`: select via that predicate,
   reject an unknown group ID, and deduplicate by block ID.

**Review gate:** list sample selectable groups and their block IDs, especially
for first-year mixed languages, shared lectures, Erasmus, and master's labs.
Do not proceed on unreviewed guesses about membership.

### 3. Calculate the actual meeting dates

8. `candidateDates(block, calendar) ([]LocalDate, error)`: term-bounded dates
   for observed frequencies (`co_tydzien`, `co_2_tygodnie`,
   `co_tydzien_pierwsza_polowa`, `co_tydzien_druga_polowa`, `jednorazowo`,
   `konkretna_data`, `pierwszy_tydzien_pazdziernika`) and periods. Anchor A/B
   to `parityAnchor` and confirm the meaning of 0/1 against the oracle.
   Handle date-specific blocks before general weekly iteration. No event for
   an unscheduled block; `allowedDates` alone does not make it scheduled.
9. `applyCalendarExceptions(block, dates, calendar) ([]LocalDate, error)`:
   closures, partial closures overlapping the event's time, `excludedDates`,
   `allowedDates`, half boundary, regular start and replacement days. A
   replacement day maps an actual date to the academic day being replaced;
   verify how that affects weekday **and parity** with fixture tests. Define
   rule precedence using observed examples, not intuition.
10. `occurrencesForBlock(block, calendar, warsaw) ([]Occurrence, error)`:
    combine dates with minutes/duration to produce local start/end instants
    and an identity based on `block.ID + actual date`. Convert to UTC for
    comparison; use IANA `Europe/Warsaw` to cover DST. Keep source block ID
    and room/campus/teacher/group metadata for downstream display and sync.
11. `occurrencesForGroup(state, groupID) ([]Occurrence, error)`: reuse
    `blocksForGroup` and `occurrencesForBlock`; avoid a second copy of rules.
12. `compareWithPublishedEvents` (**test-only**): compare multisets of
    `(blockId, UTC start, UTC end)` from the fixture against
    `calendarExport.events`. Check missing/extra occurrences, not just counts.
    Investigate every mismatch. The oracle may reflect publication notices or
    policy outside `state`; if so document the exception and ask whether to
    expand the production inputs or accept the difference. Do not quietly
    consume `calendarExport` in production to conceal a missing rule.

**Review gate:** tests for parity, DST, winter break, partial closures,
replacement days, one-offs, exclusions, semester halves, and parked blocks;
summarize remaining oracle mismatches.

### 4. Replace cache, RPC, database, and UI model

13. `buildImportedTimetable(state) (ImportedTimetable, error)`: produce
    groups, selected blocks and/or occurrences and room/location data as a
    **single immutable value**. Avoid sequentially publishing the four
    separate atomic pointers in `backend/types/cache.go` while a user sync can
    read an inconsistent mix. Decide which parts of the Excel legend/building
    resolver survive when `state.rooms` replaces legend data; preserve a raw
    room/campus location even when enrichment is unavailable.
14. `refreshTimetable(ctx) (changed bool, error)`: fetch → decode → validate →
    compare a fingerprint of the source **state** → build → atomically publish
    and persist. A bad response leaves the previous cache unchanged. Consider
    polling cadence (`MINUTES_BETWEEN`) and log change/validation reasons.
15. Change `backend/schema.sql`, `backend/queries.sql`, generated SQL, and
    `proto/pkplaner/v1/pkplaner.proto`: use `chosen_group_id TEXT` (nullable)
    instead of `chosen_column INTEGER`; carry a string group ID in selection,
    user info, and auth responses. Regenerate Go/TS clients with `buf generate`
    and SQL bindings with the project's SQL generation workflow. Update
    `backend/webServer/handler/columns.go`, `classes.go`, `backend/main.go`,
    and `frontend/src` group picker, user card, and services. Rename
    column-oriented names where practical; an unset group is represented by
    null/empty consistently, not legacy `-1`/`0`. Check for nil chosen IDs in
    background sync and return an error for invalid chosen IDs.

**Review gate:** user can pick a group and see its correct classes; no live
Google writes from the new importer yet.

### 5. Sync actual occurrences to Google Calendar

16. `eventForOccurrence(occurrence, override)`: create a dated event with
    stable private identity, `added_by`, block ID and date. Key per-class
    overrides on block ID; apply hide/time-shift/transparent ghost across its
    occurrences. Distinguish real and ghost identities. Preserve local date
    when shifting time; make update semantics explicit if start/end change.
17. `buildDesiredEvents(group, overrides)`: generate the entire desired set
    with duplicate-ID protection. Preserve useful descriptions, location and
    building links where available. Do not use the old class summary/time
    hash as an occurrence ID: a changed room or teacher should update the
    corresponding dated event, not accidentally leave an old copy.
18. `getCurrentImportedEvents` and `syncEventDiff`: enumerate **all** relevant
    Google Calendar pages within the managed time range, filter by our
    `added_by`, compare stable identities **and event contents**, and add,
    update, delete as needed. The existing `getCurrentEvents` only reads one
    page and has a rolling `TimeMin`, so it needs deliberate redesign for a
    semester of dated events. Never delete on an invalid/empty import; avoid
    flagging a user as synced when writes failed. Test the diff and a dry run
    before any live write.

**Review gate:** expected add/update/delete counts on an initial and a
changed snapshot; check legacy-event cleanup behavior separately.

### 6. Explicit cutover and cleanup

19. Point config default at the snapshot URL; update README and deployment
    instructions. Back up the deployed database and choose a **new empty DB**
    at cutover. Users reauthorize/reselect; explicitly decide whether and how
    to remove legacy events from their calendars, since a fresh DB does **not**
    automatically remove Google Calendar events created by the old version.
20. Once the new import is tested end-to-end, remove obsolete code under
    `backend/extractor`, Excel-dependent `backend/cmd` tools, XLSX-only state
    paths/stats, relevant dependencies, and `libreoffice-calc` in `Dockerfile`.
    Keep unrelated building resolution only if adapted/tested against JSON
    rooms. Update docs and run backend tests and frontend build.

## Existing code to study first

- `backend/main.go`: fetch/update/cache lifecycle and user sync ticker.
- `backend/extractor/downloader.go`, `extractor.go`: code being replaced.
- `backend/types/class.go`, `timetable.go`, `cache.go`: weekly-class and
  column-oriented assumptions; event hash and cache publication.
- `backend/calendar/calendar.go`, `differences.go`: Google event listing/diff,
  overrides and ghosts.
- `backend/webServer/handler/columns.go`, `classes.go`, `overrides.go`;
  `backend/schema.sql`, `queries.sql`; `proto/pkplaner/v1/pkplaner.proto`;
  `frontend/src/routes/Settings.tsx`.

Do not touch real user calendars or remove production data while developing or
running tests. Deliver the review gates in order, function by function.
