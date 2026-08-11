# Building Place Resolution Plan

## Goal

Convert classroom/location text from the timetable Excel file into a reliable Google Maps place link for the associated Politechnika Krakowska building.

## Resolution flow

1. Read the raw classroom/location text from the timetable.
2. Match and correct each classroom part against the classroom/location legend while creating the timetable, so later calendar-event lookups only need normalized exact cache keys.
3. Correct spelling and formatting only when the match is sufficiently confident.
4. Use the matched legend entry's canonical building name.
5. Search Google Text Search with the building name and `Politechnika Krakowska` as context.
6. Validate the returned candidates using result count, building-name appearance in the place name or address, and Politechnika Krakowska context.
7. Cache confirmed mappings and reuse them for future events.

## Failure behavior

If the classroom cannot be matched to the legend, the fuzzy match is ambiguous, or Google returns no sufficiently credible place, do not insert a place link. Keep the original classroom/location text and log the case for review.

Compound locations should be resolved per physical classroom/building part. Preserve the original location text in the calendar event and add one Google Maps link per unique physical building to the event description, separated by HTML line breaks. `ZDALNIE` should not produce a place search or link.

For now, treat classroom numbers as globally unique between buildings. During legend import, detect collisions and log them for review instead of silently replacing an existing mapping.

## Cache and refresh policy

- Use an in-memory legend mapping plus a persistent cache so Google place data is stored once per building rather than repeated for every classroom:

  ```go
  type PlaceCache struct {
  Classrooms map[string]string        // normalized classroom/alias -> building key
  Buildings  map[string]BuildingInfo  // building key -> Google place data
  }

  type BuildingInfo struct {
      CanonicalName string
      PlaceID       string
      DisplayName   string
      Address       string
      Confidence    float64
      VerifiedAt    time.Time
      ExpiresAt     time.Time
  }
  ```

- Map each normalized classroom or alias to a canonical building key.
- Store the Google `place_id`, name, address, confidence, and validation timestamps once in the building record.
- Use the canonical building name from the Excel legend as the initial building key; once confirmed, treat the Google `place_id` as the durable place identity.
- Re-evaluate classroom-to-building mappings when a new Excel legend is imported or at the start of an academic term.
- Refresh Google place details approximately every 60–90 days.
- Search for a place only when a building has no place ID, its legend mapping changes, or an administrator invalidates the existing mapping.
- Refresh stale place details during timetable import rather than on every application startup.
- If a refresh fails, retain the last known valid place and log the failure. Remove an existing place link only after a confirmed invalid result or manual rejection.

### Persistent database design

The existing building, classroom, and join tables may be replaced because the current schema has not yet been migrated to production. Persist only the Google place cache; the classroom-to-building mapping is rebuilt from the current Excel legend on every import:

```sql
CREATE TABLE building_places (
    building_key TEXT PRIMARY KEY,
    canonical_name TEXT NOT NULL,
    place_id TEXT,
    display_name TEXT,
    address TEXT,
    confidence REAL NOT NULL DEFAULT 0,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP
);

```

This stores Google data once per building. The in-memory legend mapping connects normalized classroom aliases to `building_key`; `place_id` is the durable Google identity.

## Review queue

Store unresolved or low-confidence cases in a durable JSON review file rather than building an admin panel initially. Each item should include the original text, normalized parts, corrected legend match, building name, confidence score, Google candidates, validation reasons, status, timestamps, and any administrator decision. Allow the file to be edited by hand; accepted decisions should be cached so the same case does not require repeated review.

Send best-effort webhook notifications when a review item is created or an error requires attention. Webhook delivery failures should only be logged and must not prevent timetable processing.

## Implementation notes

- Move the matcher out of `backend/cmd/matchclassrooms` into a reusable package.
- Keep exact, fuzzy, ambiguous, and unmatched outcomes distinct.
- Add a score margin between the best and second-best fuzzy candidates before accepting a correction. If a tie remains, choose the first candidate in deterministic sorted order and log the tie.
- Use stricter fuzzy thresholds for short classroom names and progressively more permissive thresholds for longer names. Keep these thresholds centralized and covered by matcher tests.
- Preserve the canonical legend entry so its associated building name can be used in the Google query.
- Use Google Places Text Search with the canonical building name and `Politechnika Krakowska` as context, inspect multiple candidates, and store only a sufficiently validated candidate.
- Treat transient Google/API failures as structured log errors. Persist deterministic unresolved cases in the JSON review file.
- Prefer conservative behavior: an unresolved classroom is better than an incorrect map link.

## Current implementation status

### Stage 1 complete — in-memory legend resolution

- `backend/classroommatch` owns deterministic location splitting, normalization, exact matching, conservative fuzzy correction, ambiguity handling, and collision detection. `cmd/matchclassrooms` is now only a diagnostic client of that package.
- Timetable construction builds the classroom/building cache from the imported Excel legend, then stores resolved normalized classroom keys on every `Class`. The original classroom text remains unchanged for calendar display.
- The cache maps normalized classroom aliases to the legend's canonical building name. It supports room qualifiers, `bis`, compound locations, `ZDALNIE`, and building-acronym aliases such as `s. SJO`, `s. 152 SJO`, and `s. 139 SJO`.
- Calendar-event building lookup now uses only the precomputed exact keys. The legacy seeded database mapping is no longer used by the runtime path, so no legacy map links are produced.
- Structured `classroom_resolution.*` logs expose index contents, corrections, unresolved cases, and per-class `location -> building` mappings for validation.

### Stage 2 complete — persistent Google Places resolution

- The legacy seeded schema has been replaced by the `building_places` cache.
  Classroom-to-building mappings are rebuilt from the imported legend. This is
  a fresh database version: deployments
  remove the prior `state/database.db` rather than migrate it.
- Cache construction at startup and after a changed timetable import syncs the
  legend mappings, resolves missing buildings with Google Places Text Search,
  refreshes confirmed details after 90 days, and hydrates confirmed map links.
- Candidate validation is name-led: canonical building-name evidence is
  required, PK and Kraków context increase confidence, and the result must
  score at least 0.70 with a 0.15 lead over the next result.
- `state/building_place_review.json` is a durable hand-editable queue for
  unresolved classrooms and uncertain Google candidates. An optional
  `BUILDING_REVIEW_WEBHOOK_URL` receives notifications for newly created items.
- Existing calendar events are left unchanged. Confirmed map links appear when
  future events are created.
