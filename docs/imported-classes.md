# Classes from other groups

Settings → **Dodaj zajęcia** opens a group picker and searchable class list.
The server offers only groups in the same academic year and study programme as
the user's selected group. Semester pairs 1/2, 3/4, etc. define the year; category
prefixes and specialization suffixes must match. Categories without a recognized
semester are excluded rather than guessed.

## Storage and synchronization

An inclusion is a normal override with a nonzero `source_column`. Its `class_hash`
identifies the original published class. Hide, time-shift, and transparency flags
remain available. An inclusion with otherwise empty flags must **not** be deleted.
Resetting an imported class removes the inclusion; resetting a base-group class
still removes only its customization.

`calendar.UserTimetable` combines the base group with eligible referenced classes,
deduplicates by hash, and never modifies the shared cache. Both authenticated
`GetClasses` and Google Calendar synchronization use this composition. Public
`GetTimetable` continues to return the published group only, without additions or
user overrides.

References are resolved against the current timetable on each read/sync. Missing
classes and references outside the selected group's academic year are omitted.
Changing the base group therefore cannot import classes from a previous year.
Like existing overrides, a reference whose source class hash changes needs to be
selected again; no stale snapshot of a class is retained.

## Deployment

Rebuild and restart the backend for the new RPC and protobuf field. The schema adds
`overrides.source_column` with default `0`. As with `users.calendar_id`, existing
databases are not upgraded in place — provision a fresh database. Adding or removing
a class triggers that user's sync immediately; it also reaches Google Calendar on the
next scheduled publication sync.
