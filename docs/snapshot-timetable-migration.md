# Hybrid timetable importer — implementation plan

Agreed direction: 2026-10-04. This replaces the full JSON migration plan.

## Goal

Keep Excel for class details and group membership. Fetch JSON alongside it
to determine when each class actually meets:

`https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php`

Preserve the existing group picker and overrides. Excel already distinguishes
regular, mixed-language, and elective columns. The extractor merges language
and elective classes into regular groups of the same category, allowing users
to hide the options they do not attend. Do not reconstruct that membership
from JSON cohorts or add separate language/elective selectors.

Use JSON `state.blocks` for each class's scheduling rules and
`state.academicCalendar` for global calendar rules. `calendarExport.events`
is an offline test oracle only, not a production input.

## First step: prove matching works

Before changing the importer, compare a current Excel workbook with a captured
JSON response offline. Ensure they describe the same term and publication as
closely as possible; the two sources can be updated at different times.

Index JSON blocks by subject:

```go
map[string][]snapshot.Block
```

For each extracted Excel class, narrow candidates by activity, weekday, and
start time. Use end time, teacher, and audience/group information to resolve
remaining candidates where their representations are understood. Any text
normalization should be small and based on observed differences.

Keep matched block IDs with imported classes so sync does not repeat matching.
The same block may appear in several Excel groups; that is expected.
Subject alone is not an identity. Records sharing a subject and time can have
different activities or scheduling rules.

## JSON metadata model and cache

Treat extracted JSON scheduling metadata like the existing legend: build it
once during import, hold it in the cache behind an
`atomic.Pointer[SnapshotMetadata]`, and pass the loaded reference through to
matching, occurrence generation, and calendar sync. `SnapshotMetadata` is a
proposed name, not a required abstraction framework.

The value should contain the academic calendar and block lookups needed by
the importer, initially by subject for matching and by ID for scheduling.
Keep it immutable after publication. Each operation loads the pointer once
and passes that same value through; it should not refetch JSON or reload the
atomic pointer for each class. Atomic pointer access belongs in the cache;
downstream functions receive an ordinary `*SnapshotMetadata`.

Build a replacement value on refresh and publish it only after the combined
Excel/JSON import succeeds. The user chose a separate atomic snapshot field in
Cache alongside the timetable, accepting independent loads. Do not introduce
a paired generation wrapper or mutex. Cache currently retains *snapshot.State.

Report unique matches, ambiguous matches, and unmatched records with examples.
In particular, check languages, electives, shared lectures, and master's labs.
Do not invent mappings between labels such as Excel `SIRR` and JSON `SIR`.
Only unique matches receive Class.SnapshotBlockID. Unmatched or ambiguous
classes retain an empty ID and produce a warning; they do not abort import.
Define calendar behavior for these classes before enabling snapshot scheduling.

## Implementation after the matching review

1. Move the active extractor from `backend/legacy/extractor` back to
   `backend/extractor`, updating imports. The older heuristic text parser can
   remain in `backend/legacy`. Keep the working Excel download/conversion flow.
2. Fetch and decode JSON alongside Excel using the existing `backend/snapshot`
   code. Trust the publisher's field values; check basic compatibility
   (supported schema/publication mode and nonempty data) where the combined
   timetable is built. No exhaustive per-field validator is needed.
3. Build the Excel timetable as today, then attach matched block IDs and pass
   the cached JSON metadata reference through as described above. Keep
   existing column selection and class override keys; no fresh
   database or string group-ID migration is required for this approach.
4. Add one occurrence generator using the matched block and academic calendar.
   Parse dates/times as needed and return errors if they cannot be used.
   Introduce smaller helpers only when they make the implementation clearer.
5. Adapt Google Calendar event generation and sync to actual dated meetings.
   Keep the existing hide, time-shift, and transparent ghost behavior, applying
   each class override across that class's occurrences.
6. Refresh Excel and JSON together even when the Excel file has not changed:
   calendar metadata can change independently. Build a complete combined result
   before publishing it. Fetch/decode/matching/build failures retain the
   previous working cache and must not trigger deletions from an empty result.

## Scheduling behavior

- Weekly and A/B frequency comes from each matched block. A/B classes use
  `co_2_tygodnie` plus `teachingWeekParity`; the committed fixture confirms
  `0 = A`, `1 = B`. Use `parityAnchor`, not a hardcoded October date or ISO
  week parity. In the current calendar it is 2026-09-28, with the next Monday
  beginning B week.
- Treat blocks independently. Do not infer a complementary A/B class from
  another subgroup's slot or assume both subgroups attend equally often.
- `co_tydzien_pierwsza_polowa` and `co_tydzien_druga_polowa` refer to halves of
  the semester, not halves of an exercise group. Apply the supplied period and
  `halfBoundary`. In the fixture, lecture `a2558-1` ends on 18 November and
  projects `a2560-1`/`a2560-2` begin on 25 November around the 23 November
  boundary. This is evidence of publisher behavior, not a permanent date rule.
- Apply term bounds, closures, overlapping partial closures, replacement days,
  regular stationary start, allowed dates, and excluded dates. Determine their
  precedence and replacement-day weekday/parity behavior from oracle examples.
- Skip `jednorazowo` and `pierwszy_tydzien_pazdziernika` blocks, per the user's
  decision. Retain `konkretna_data`. Keep intentional omissions
  separate from unexplained oracle mismatches.
- Missing day/start means parked: do not guess an occurrence from allowed dates.
- Use `Europe/Warsaw` for local meeting times and DST.

## Calendar sync changes that remain necessary

Agreed on 2026-10-04: create one dedicated "PlanerPK" secondary calendar per
user, persist its ID, and sync timetable events there rather than to "primary".
The user confirmed there are no existing importer events, so no migration of
old events is needed.

Use one recurring series per matched class/block per user. Weekly and A/B
classes use RRULE with semester or half-semester bounds; use EXDATE for skipped
meetings and RDATE for added replacement dates. Concrete-date classes remain
single events. Generate and verify actual meeting dates first, then encode
the verified schedule as recurrence data. Changed individual meeting times
may require instance exceptions.

Use block-based series identity; distinguish real and ghost series and identify
individual exceptions by their original occurrence. Keep override identity
separate so one override applies across a series. Compare event contents and
recurrence data as well as identity when synchronizing.

List all relevant Google Calendar pages over the managed term range. The current
single-page, rolling-TimeMin listing is insufficient for a semester of meetings.
Adapt listing to recurring series and exceptions rather than treating every
expanded instance as an independent event. Do not mark a sync successful when
writes fail.

## Verification and scope

Use captured, sanitized Excel/JSON fixtures for tests. The existing JSON fixture
was captured on 2026-10-02 and contains 369 blocks and 4,057 exported events;
these are fixture counts, not importer requirements. Compare dated occurrences
by block ID and UTC start/end, not just totals. Review matching coverage,
A/B weeks, term halves, holidays, replacements, exclusions, DST, and overrides.
Review a dry-run event diff before any live calendar writes.

Keep existing room/building enrichment where it still works with Excel.
Defer broad renaming, generic matching frameworks, extensive validation, and
fallback-persistence infrastructure. Retain the current persisted Excel flow;
startup must obtain usable matching JSON or fail rather than publish guessed
dates. A persisted JSON fallback can be added later if offline startup is needed.

Existing snapshot DTOs, bounded fetch/decode functions, and the sanitized oracle
fixture remain useful. The standalone validation/persistence implementation
from the abandoned full JSON migration has been removed. No importer cutover,
package move, production data removal, or live calendar writes are part of this
planning update.
