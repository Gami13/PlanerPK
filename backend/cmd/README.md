# ClassText Scripts

Scripts for extracting and processing classText samples from the timetable Excel file. Run all commands from the `backend/` directory.

## Snapshot subject index

Fetch the online snapshot and inspect its subject index (without importing or syncing):

```sh
go run ./cmd/indexsnapshot
```

The command takes no parameters and uses the published snapshot endpoint. Output is JSON; index totals go to stderr.

## Pipeline

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


