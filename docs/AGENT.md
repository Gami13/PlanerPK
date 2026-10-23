# Timetable importer handoff

## Event construction follow-up

`types.Class.ToEvent` now accepts a neutral `*types.ClassSchedule` with resolved
start/end times and recurrence rules. Existing live callers pass nil, retaining
the old next-weekday behavior temporarily. Remove that fallback when snapshot
scheduling is connected, after agreeing how unmatched classes should behave.
Do not retain next-weekday scheduling as an implicit unmatched-class policy.

Read `docs/snapshot-timetable-migration.md` for the current plan. The earlier
full JSON replacement plan has been superseded by a hybrid importer.

## User preferences and agreed direction

- Keep the implementation small. The user rejected excessive validation,
  builder callbacks in persistence, platform-specific replacement helpers,
  and treating every proposed function in the old document as mandatory.
- Implement one small, reviewable step per turn, ideally under 100 changed
  lines including tests. Explain the step before editing; discuss splitting
  larger steps. Run relevant checks, summarize the diff, then stop for review.
  Do not proceed to another step until the user says so. Commit only on request.
- Inspection commands take no URL/subject flags. Use the published endpoints.
  Keep successful match output short; retain details for ambiguous/unmatched rows.
- Trust the university's published field values. Keep basic compatibility
  checks and handle parsing errors where values are used; do not restore an
  exhaustive `validateState` implementation.
- Excel owns group membership and class details. JSON owns scheduling rules
  and actual dates. Keep the existing group picker and override behavior.
- Extract JSON metadata once per import, cache it atomically like the legend,
  and pass a loaded reference through downstream functions. Keep the
  timetable paired with its metadata generation. See the plan's cache section.
- Skip frequency `jednorazowo` in calendar sync: the user relies on Teams
  announcements. This decision does not extend to `konkretna_data`.
- Clarifying questions are welcome when needed. Do not start a broad rewrite
  while the matching approach remains unproven.

## What the existing code actually does

The active Excel extractor currently lives in `backend/legacy/extractor`.
Moving it back to `backend/extractor` is planned but has not happened.
The older heuristic text parser in `backend/legacy/parser.go` can stay there.

`structured_parser.go` retains subject, activity, teacher, audience, room,
weekday, and times. It preserves all trailing scheduling notes as raw text in
`types.Class.AdditionalInfo`; it does not interpret them as date rules.

`extractor.go` identifies supplemental Excel columns by labels containing
`językowe` or `obieralne`. It merges their classes into regular groups of the
same category and excludes these supplemental columns from the picker.
Users hide unwanted alternatives using overrides. Preserve this behavior.

`types.Class.ToEvent` currently places a class on its next weekday. It does
not interpret A/B or semester-half notes. Calendar listing currently reads
one page with a rolling TimeMin; dated semester imports require adaptation.

## Findings and cautions

- The user's screenshots were from Excel, not the public web page. Do not
  repeat the earlier mistaken claim that they showed web rendering.
- The live snapshot was inspected on 2026-10-04. The sampled semester-3
  language, business communication, and negotiations blocks did not have
  elective/mixed-group markers; `state.legacy` was empty. First-year language
  blocks in the fixture do have `studentGrouping.kind = mixed_language`.
  The JSON membership metadata is insufficient to simply replicate Excel's
  supplemental-column behavior. This is why the hybrid was chosen.
- Do not identify languages by substring `Język`: that also matches
  `Języki i paradygmaty programowania`. No subject-based elective list has
  been implemented or adopted as the hybrid membership policy.
- `co_2_tygodnie` plus parity is A/B. The fixture confirms `0 = A`, `1 = B`,
  anchored at 2026-09-28. Use calendar metadata rather than hardcoding dates.
- `co_tydzien_pierwsza_polowa`/`co_tydzien_druga_polowa` refer to semester
  halves. Oracle block `a2558-1` has lectures through 18 November, while
  projects `a2560-1`/`a2560-2` begin 25 November. The supplied half boundary
  is 23 November. These are fixture observations, not permanent dates.
- Treat blocks independently; do not invent complementary A/B attendance.
- JSON cohort text sometimes says `SIR`, but examined master's blocks also
  supply `SIRR` academicGroupLabel/studentGrouping labels. Compare demonstrated
  explicit metadata; do not invent fuzzy mappings between cohort names.
- `calendarExport` is a test oracle only. Production uses `state`.

## Current state and next step

Completed and committed on 2026-10-04:

- `8780c57` direct XLSX downloader. Default TIMETABLE_URL is
  `https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/download.php`.
  Uses Content-Disposition filename and copies XLSX without conversion.
  HEAD lacks Last-Modified, so it currently refreshes every polling interval.
- `0ec3163` removed XLS conversion, old HTML discovery/helpers/tests, their
  unused Go dependencies, and LibreOffice dependencies from Docker/Nix.
- `c6aaead` subject index and online inspection command; `67f71a9` made it
  `snapshot.SubjectIndex` with `NewSubjectIndex` and `Candidates(types.Class)`.
  Matching belongs in snapshot, not a separate timetablematch package.
- `623dcd3` comparison command, metadata filters, subject fallback, and refactor.
  `Block.matchesAudience` handles group comparison; the slot filter uses an
  early continue. Existing snapshot Fetch/Decode are exported and reused.

Run commands from backend:
`go run ./cmd/indexsnapshot` dumps the live subject index as JSON.
`go run ./cmd/matchsnapshot` compares the current workbook with live JSON;
DATA_DIR defaults to ../data. It does not load OAuth credentials or write
calendars. Unique rows show group, subject/activity, block ID; subject fallback
shows both names. Other rows retain full details. Totals go to stderr.

Current workbook: data/current_timetable/Plan-ST-2026-2027.xlsx. Latest live
comparison: 365 blocks, 77 subjects, 847 unique Excel records, 0 ambiguous,
0 unmatched. Records repeated across columns and merged supplemental classes
count separately. These are observed counts, not expected totals or proof
that every future publisher change is handled. Fixture remains 369 blocks.

Matching tries exact subject plus activity, weekday, start time, then filters:
- Trimmed teacher matches teacher or teacherDisplay; missing names preserve
  candidates in the ordinary path.
- Numeric GL/C/gr./elective audiences compare GroupNo or cohort suffix;
  CY/DS/SIRR exercise labels compare academicGroupLabel exactly.
- First-line A/B notes compare co_2_tygodnie parity (0=A, 1=B).
  First-line DD.MM.YYYY notes compare Date. Other notes remain uninterpreted.
- Only when ordinary matching returns zero candidates, search all subjects
  using the same checks, requiring a known matching teacher. Return all
  candidates; never choose arbitrarily. Fallback results are sorted by block ID.
  This resolved two shortened names (cryptography and advanced ML/deep learning)
  without aliases or fuzzy matching. Unknown metadata preserves ambiguity.

End time/room are not matching filters yet. Compound A/B notes and semester-half
notes are not parsed here. The diagnostic reports IDs but does not attach them
to imported classes. Runtime still uses Excel and next-weekday calendar events;
no JSON cache integration, occurrence generator, importer cutover, or live
calendar writes have happened. Abandoned validation/persistence code stays removed.

Latest checks passed: go test ./snapshot ./cmd/matchsnapshot and git diff --check;
live comparison stayed 847/0/0 after audience extraction. Full backend suite
passed during XLS removal; Docker/Nix builds were not run. Go cache/network
access can require escalation; rerun denied commands with required permissions.

Next: use the matching proof to start the plan's runtime implementation in a
small reviewed step. The plan lists moving legacy/extractor back to extractor
first; assess its diff size and discuss a split before a large package move.
Then build combined Excel/JSON metadata and retain matched block IDs. Publish
the timetable and its metadata as one generation; sync must not rematch.
Capture a sanitized paired fixture before claiming offline regression coverage.
Migration docs remain uncommitted intentionally; do not bundle them into code
commits unless requested. Re-read git status before continuing.

Do not require new string group IDs, a fresh database, or discarding existing
users/overrides: those belonged to the abandoned full replacement plan.
Do not touch real user calendars or remove production data during development.

## Latest decisions and implementation (2026-10-04)

This update supersedes the older runtime status and next-step notes above.
The extractor move, snapshot.Load compatibility checks, separate cache package,
atomic snapshot cache field, and class ID attachment are committed through
ece5291. Matching runs during extractor finalization; only unique matches get
SnapshotBlockID. Missing or ambiguous matches warn and leave the ID empty.
Class hashes and calendar behavior remain unchanged. The user chose independent
atomic timetable/snapshot fields, rejecting a generation wrapper and mutex.

Calendar target: one "PlanerPK" secondary calendar per user with a persisted ID,
and one recurring series per matched class/block. Use exclusions and added dates
for schedule exceptions; concrete-date meetings remain single events. The user
confirmed there are no existing importer events, so skip old-event migration.
Next: build and verify actual meeting dates in snapshot in small reviewed steps,
then encode recurrence and adapt sync. No occurrence generator exists yet.
The user also chose to omit pierwszy_tydzien_pazdziernika, alongside
jednorazowo; concrete dates and semester halves must be supported.
