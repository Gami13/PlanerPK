# Building place resolution

## Goal

Convert timetable classroom text into reliable Google Maps links for the
associated Politechnika Krakowska buildings while preserving the original
location text shown in calendar events.

## Runtime flow

1. Import the classroom-to-building entries from the Excel legend.
2. Build separate exact-classroom and building-acronym indexes.
3. Parse and merge timetable slots, split slash-delimited subjects into their
   final classes, and resolve each final class location once.
4. Store normalized classroom keys on the class. Calendar-event generation
   performs exact lookups only and emits one link per unique building.
5. Synchronize the legend's unique buildings into `building_places`, resolve
   missing places, refresh expired details, and hydrate confirmed links.

`ZDALNIE` is ignored. Compound physical locations are resolved one part at a
time, while numeric room forms such as `1/15` remain a single part.

## Classroom matching

Exact classroom aliases come from legend shorthands, including normalized
forms with room qualifiers and `bis`. Building aliases are indexed separately
and are derived once per unique building only when a legend name contains an
uppercase acronym after `budynek`, for example `SJO`.

Aliases that point to different buildings are excluded rather than resolved by
insertion order. An exact classroom alias takes precedence over a compatible
same-building acronym alias. Conflicting acronym aliases are also excluded.

If no exact alias matches, the matcher may correct a classroom using Unicode-
aware normalization and Levenshtein similarity. Names of three characters or
fewer are never corrected. The minimum scores are 0.90 for four to five
characters, 0.82 for six to eight, and 0.75 for longer names. The winning score
must lead the second-best score by at least 0.10. A close result is logged as a
warning; ties are not accepted. Unmatched and ambiguous locations retain the
original text but receive no map link.

## Persistent place cache

`building_places` stores one record per canonical legend building. A record
contains the building key and name plus an optional Google place ID, display
name, address, confidence, verification time, and expiry time. Classroom maps
are rebuilt in memory on each import and are not persisted.

Missing records are searched with the canonical building name and
`Politechnika Krakowska` context when `GOOGLE_PLACES_API_KEY` is configured.
Candidate validation requires building-name evidence, a score of at least
0.70, and a lead of at least 0.15 over the next candidate. Confirmed details
expire after 90 days and are refreshed during timetable import. A failed
refresh retains the last confirmed place.

## Failure behavior and diagnostics

Google/API failures do not prevent a timetable import. The application keeps
raw classroom text and omits unconfirmed links. Structured
`classroom_resolution.*` logs report collisions, corrections, ambiguity, and
unmatched input. Ambiguous and unmatched classroom results are emitted at warn
level, while rejected Google Place candidates are emitted at error level with their building key,
confidence, candidate count, and validation reasons. Final-class debug logs
include the resolved classroom keys and building names without dumping the full
index.
