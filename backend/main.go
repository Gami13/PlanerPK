package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/cache"
	"github.com/woodpkrs/PlanerPK/calendar"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/webServer"

	"github.com/woodpkrs/PlanerPK/config"

	"github.com/woodpkrs/PlanerPK/extractor"
	"github.com/woodpkrs/PlanerPK/utils"
)

//go:embed schema.sql
var ddl string

func initializeXlsxInfo() {
	filename := config.Config.TimetableStats

	_, err := os.Stat(filename)
	if err == nil {
		slog.Info("Stats file already exists")
		return
	}
	if !errors.Is(err, os.ErrNotExist) {
		utils.NotifyError("Failed to check if stats file exists", slog.Any("error", err))
		panic("Failed to check if stats file exists: " + err.Error())
	}

	jsonText, err := json.Marshal(types.TimetableFileStats{})
	if err != nil {
		utils.NotifyError("Could not marshal json", slog.Any("error", err))
		panic("Could not marshal json")
	}
	err = os.WriteFile(filename, jsonText, 0o644)
	if err != nil {
		utils.NotifyError("Could not init stats file", slog.Any("error", err))
		panic("Could not init stats file")
	}
	slog.Info("Stats file initialized")
}

// Function is free to panic as resources are required for app to function
func initializeApp() (queries db.Querier) {
	config.Load()

	err := utils.InitializeLogging(config.Config.LogsDir, config.Config.LogLevel)
	if err != nil {
		utils.NotifyError("Failed to initialize logger", slog.Any("error", err))
		panic("Failed to initialize logger: " + err.Error())
	}
	slog.Info("Configuration loaded",
		slog.String("data_dir", config.Config.DataDir),
		slog.Bool("building_locations_enabled", config.FEATURE_FLAGS.BuildingLocationsEnabled),
		slog.Bool("google_places_api_key_set", config.Config.GooglePlacesAPIKey != ""))

	database, err := sql.Open(config.DatabaseDriver, config.Config.DatabaseSource)
	if err != nil {
		utils.NotifyError("Failed to open database", slog.Any("error", err))
		panic(err)
	}

	_, err = database.Exec(ddl)
	if err != nil {
		utils.NotifyError("Failed to execute DDL statements", slog.Any("error", err))
		panic(err)
	}

	queries = db.New(database)
	slog.Info("Database initialized successfully")

	initializeXlsxInfo()
	return queries
}

func buildNewCache(queries db.Querier) ([]types.ColumnGroup, *legend.Legend, *buildings.Buildings, types.Timetable, *snapshot.State, error) {
	newSnapshot, err := snapshot.Load(http.DefaultClient, "https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	file, err := extractor.OpenCurrentTimetableFile()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer extractor.CloseFile(file)
	sheetName, err := extractor.ResolveTimetableSheet(file)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	layout, err := extractor.DiscoverTimetableLayout(file, sheetName)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	allGroups, err := extractor.BuildAvailableGroups(file, sheetName, layout)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	newLegend := extractor.ExtractLegend(file, sheetName)
	var newBuildings *buildings.Buildings
	if config.FEATURE_FLAGS.BuildingLocationsEnabled {
		newBuildings = buildings.BuildBuildingsFromLegend(newLegend)
		if err := buildings.ResolveBuildingPlacesForImport(context.Background(), queries, newBuildings); err != nil {
			return nil, nil, nil, nil, nil, err
		}
	} else {
		slog.Info("Building location resolution is disabled")
	}
	index := snapshot.NewSubjectIndex(newSnapshot.Blocks)
	newTimetable := extractor.BuildTimetable(allGroups, file, sheetName, newLegend, newBuildings, layout, index)
	newGroups := extractor.SelectableGroups(allGroups)
	return newGroups, newLegend, newBuildings, newTimetable, &newSnapshot, nil
}

func updateTimetableAndCache(queries db.Querier, cache *cache.Cache) (bool, error) {
	needsDownload, fileURL, err := extractor.DoesTimetableNeedUpdating()
	if err != nil {
		return false, err
	}
	if needsDownload {
		// Download new timetable
		slog.Info("Getting new timetable")
		stats, err := extractor.DownloadTimetable(fileURL)
		if err != nil {
			slog.Error("Failed to download the new timetable", "error", err)
			return false, err
		}
		slog.Info("Processing new timetable")
		_, err = extractor.ProcessTimetable(stats)
		if err != nil {
			return false, err
		}

	}
	// JSON can change independently of Excel. Rebuild before publishing either.
	newGroups, newLegend, newBuildings, newTimetable, newSnapshot, err := buildNewCache(queries)
	if err != nil {
		return false, err
	}
	workbookHash, err := extractor.CurrentTimetableHash()
	if err != nil {
		return false, err
	}
	previous := cache.Snapshot()
	// Trust the publisher's fingerprint; a missing hash conservatively triggers sync.
	hasChanged := cache.TimetableHash() != workbookHash || previous == nil || newSnapshot.Snapshot.SHA256 == "" ||
		previous.Snapshot.SHA256 != newSnapshot.Snapshot.SHA256
	cache.SetBuildings(newBuildings)
	cache.SetAvailableGroups(&newGroups)
	cache.SetLegend(newLegend)
	cache.SetSnapshot(newSnapshot)
	cache.SetTimetable(&newTimetable)
	cache.SetTimetableHash(workbookHash)

	return hasChanged, nil
}

func main() {
	queries := initializeApp()

	var cache cache.Cache

	_, err := updateTimetableAndCache(queries, &cache)
	if err != nil {
		utils.NotifyError("Failed to build initial cache", slog.Any("error", err))
		panic(err)
	}

	ticker := time.NewTicker(config.Config.TimetableUpdatePeriod)
	defer ticker.Stop()

	slog.Info("Starting web server", slog.String("port", config.ServerPort))

	go webServer.Start(queries, &cache)

	slog.Info("Starting timetable update ticker",
		slog.Int("interval_minutes", int(config.Config.TimetableUpdatePeriod/time.Minute)))
	pendingSync := true // Check existing users once after startup as well.
	for t := range ticker.C {
		slog.Info("Timetable update triggered", slog.Time("tick_time", t))
		pendingSync = updateAllTimetables(queries, &cache, pendingSync)
	}
}

// Return whether the global sync batch still needs retrying next poll.
func updateAllTimetables(queries db.Querier, cache *cache.Cache, pendingSync bool) bool {
	changed, err := updateTimetableAndCache(queries, cache)
	if err != nil {
		slog.Error("Failed to update timetable and cache", slog.Any("error", err))
		return pendingSync
	}
	if !changed && !pendingSync {
		slog.Info("Timetable cache unchanged; skipping scheduled calendar sync")
		return false
	}
	users, err := queries.GetAllUsers(context.Background())
	if err != nil {
		slog.Error("Failed to fetch users", slog.Any("error", err))
		return true
	}
	allTimetables := cache.Timetable()
	state := cache.Snapshot()
	failed := false

	for _, user := range users {
		if *user.ChosenColumn < 0 || !*user.IsActive {
			slog.Info("Skipping timetable update for inactive user or user without chosen column",
				slog.String("email", *user.Email),
				slog.Int64("column", *user.ChosenColumn))

			continue
		}
		slog.Info("Updating timetable for user",
			slog.String("email", *user.Email),
			slog.Int64("column", *user.ChosenColumn))
		if err := calendar.ImportTimetable(allTimetables, utils.GetToken(&user), queries, state, cache.Buildings(), *cache.AvailableGroups()); err != nil {
			slog.Error("Timetable sync failed", slog.Int64("user_id", user.ID), slog.Any("error", err))
			failed = true
		}
	}
	return failed
}
