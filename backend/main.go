package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/calendar"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/webServer"

	"github.com/woodpkrs/PlanerPK/global"

	"github.com/woodpkrs/PlanerPK/extractor"
	"github.com/woodpkrs/PlanerPK/utils"
)

//go:embed schema.sql
var ddl string

func initializeXlsxInfo() {
	filename := global.Config.TimetableStats

	_, err := os.Stat(filename)
	if err == nil {
		slog.Info("Stats file already exists")
		return
	}
	if !errors.Is(err, os.ErrNotExist) {
		panic("Failed to check if stats file exists: " + err.Error())
	}

	jsonText, err := json.Marshal(types.TimetableFileStats{})
	if err != nil {
		panic("Could not marshal json")
	}
	err = os.WriteFile(filename, jsonText, 0o644)
	if err != nil {
		panic("Could not init stats file")
	}
	slog.Info("Stats file initialized")
}

// Function is free to panic as resources are required for app to function
func initializeApp() (queries db.Querier) {
	global.Load()
	config := global.Config

	err := utils.InitializeLogging(config.LogsDir, config.LogLevel)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	slog.Info("Configuration loaded",
		slog.String("data_dir", config.DataDir),
		slog.Bool("google_places_api_key_set", config.GooglePlacesAPIKey != ""))

	database, err := sql.Open(global.DatabaseDriver, config.DatabaseSource)
	if err != nil {
		panic(err)
	}

	_, err = database.Exec(ddl)
	if err != nil {
		slog.Error("Failed to execute DDL statements", slog.Any("error", err))
		panic(err)
	}

	queries = db.New(database)
	slog.Info("Database initialized successfully")

	initializeXlsxInfo()
	return queries
}

func buildNewCache(queries db.Querier) ([]types.ColumnGroup, *legend.Legend, *buildings.Buildings, types.Timetable, error) {
	file, err := extractor.OpenCurrentTimetableFile()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer extractor.CloseFile(file)
	newGroups, err := extractor.BuildAvailableGroups(file)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	newLegend := extractor.ExtractLegend(file)
	newBuildings := buildings.BuildBuildingsFromLegend(newLegend)
	if err := buildings.ResolveBuildingPlacesForImport(context.Background(), queries, newBuildings); err != nil {
		return nil, nil, nil, nil, err
	}
	newTimetable := extractor.BuildTimetable(newGroups, file, newLegend, newBuildings)
	return newGroups, newLegend, newBuildings, newTimetable, nil
}

func updateTimetableAndCache(queries db.Querier, cache *types.Cache, isInitialRun bool) (hasChanged bool, err error) {
	hasChanged, fileURL, err := extractor.DoesTimetableNeedUpdating()
	if err != nil {
		return false, err
	}
	if hasChanged {
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
	if isInitialRun || hasChanged {
		newGroups, newLegend, newBuildings, newTimetable, err := buildNewCache(queries)
		if err != nil {
			return false, err
		}
		cache.SetBuildings(newBuildings)
		cache.SetAvailableGroups(&newGroups)
		cache.SetLegend(newLegend)
		cache.SetTimetable(&newTimetable)
	}

	return hasChanged, nil
}

func main() {
	queries := initializeApp()

	var cache types.Cache

	_, err := updateTimetableAndCache(queries, &cache, true)
	if err != nil {
		slog.Error("Failed to build initial cache", "error", err)
		panic(err)
	}

	ticker := time.NewTicker(global.Config.TimetableUpdatePeriod)
	defer ticker.Stop()

	slog.Info("Starting web server", slog.String("port", global.ServerPort))

	go webServer.Start(queries, &cache)

	slog.Info("Starting timetable update ticker",
		slog.Int("interval_minutes", int(global.Config.TimetableUpdatePeriod/time.Minute)))
	for t := range ticker.C {
		slog.Info("Timetable update triggered", slog.Time("tick_time", t))
		updateAllTimetables(queries, &cache)
	}
}

func updateAllTimetables(queries db.Querier, cache *types.Cache) {
	users, err := queries.GetAllUsers(context.Background())
	if err != nil {
		slog.Error("Failed to fetch users", slog.Any("error", err))
		return
	}
	hasChanged, err := updateTimetableAndCache(queries, cache, false)
	if err != nil {
		slog.Error("Failed to update timetable and cache", slog.Any("error", err))
	}
	allTimetables := cache.Timetable()

	for _, user := range users {
		if *user.ChosenColumn < 0 || !*user.IsActive {
			slog.Info("Skipping timetable update for inactive user or user without chosen column",
				slog.String("email", *user.Email),
				slog.Int64("column", *user.ChosenColumn))

			continue
		}
		timeSinceSync := time.Since(user.LastSynced)
		if !hasChanged && timeSinceSync < global.Config.TimetableSyncPeriod {
			slog.Info("Skipping timetable update for user (no changes detected and last sync within 3 days)",
				slog.String("email", *user.Email),
				slog.Int64("column", *user.ChosenColumn))
			continue
		}

		slog.Info("Updating timetable for user",
			slog.String("email", *user.Email),
			slog.Int64("column", *user.ChosenColumn))
		overrides, err := queries.GetOverridesForUser(context.Background(), user.ID)
		if err != nil {
			slog.Warn("Failed to fetch overrides for user, syncing without overrides",
				slog.String("email", *user.Email),
				slog.Any("error", err))
			overrides = nil
		}
		calendar.ImportTimetable(allTimetables, *user.ChosenColumn, utils.GetToken(&user), queries, overrides, cache.Buildings())
	}
}
