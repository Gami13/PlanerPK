# ClassText Scripts

Scripts for extracting and processing classText samples from the timetable Excel file. Run all commands from the `backend/` directory.

## Pipeline

```sh
go run ./cmd/collectclasstexts/
go run ./cmd/dedupclasstexts/
go run ./cmd/collapsespaces/
go run ./cmd/splitclasstexts/
go run ./cmd/splitbycount/
go run ./cmd/dumpclassrooms/
go run ./cmd/matchclassrooms/
```


