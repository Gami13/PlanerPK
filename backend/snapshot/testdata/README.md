# Snapshot fixture

`winter-2026-27.json` was captured from the timetable snapshot endpoint on
2026-10-02 for offline tests. It retains all 369 source blocks, 25 rooms,
academic-calendar rules, and 4,057 oracle event times from that capture.
These counts are fixture-specific, not importer requirements.

The fixture is **sanitized**: teacher and teacher-display values are blank;
export event UIDs, descriptions, summaries, locations, notices, wishes,
constraints, legacy data, and other unrelated publisher fields were removed.
The publisher SHA-256 was also removed because the fixture was changed.
`calendarExport` is retained only as a test oracle; production imports must
derive dates from `state`.
