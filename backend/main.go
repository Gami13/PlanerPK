package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"

	"github.com/gami13/PlanerPK/calendar"
	"github.com/gami13/PlanerPK/gen/db"
	"github.com/gami13/PlanerPK/types"
	"github.com/gami13/PlanerPK/webServer"

	"github.com/gami13/PlanerPK/global"

	"github.com/gami13/PlanerPK/extractor"
	"github.com/gami13/PlanerPK/utils"
)

//go:embed schema.sql
var ddl string

func initializeXlsxInfo() {
	filename := filepath.Join(global.JSON_INFO_LOCATION, global.TIMETABLE_CURRENT_STATS_NAME)
	filename, err := filepath.Abs(filename)
	if err != nil {
		panic("Couldn't create stats file path")
	}

	_, err = os.Stat(filename)
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
	err = os.WriteFile(filename, jsonText, 0644)
	if err != nil {
		panic("Could not init stats file")
	}
	slog.Info("Stats file initialized")
}

// Function is free to panic as resources are required for app to function
func initializeApp() (queries db.Querier) {
	global.LoadEnvVariables()

	err := utils.InitializeLogging(global.LOGS_PATH)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	slog.Info("Environment variables loaded")

	global.PrintEnvVariables()

	database, err := sql.Open(global.DATABASE_DRIVER, global.DATABASE_SOURCE)
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

func buildNewCache() ([]types.ColumnGroup, *types.Legend, types.Timetable, error) {
	file, err := extractor.OpenCurrentTimetableFile()
	if err != nil {
		return []types.ColumnGroup{}, &types.Legend{}, types.Timetable{}, err
	}
	defer extractor.CloseFile(file)
	newGroups, err := extractor.BuildAvailableGroups(file)
	if err != nil {
		return []types.ColumnGroup{}, &types.Legend{}, types.Timetable{}, err
	}
	newLegend := extractor.ExtractLegend(file)
	newTimetable := extractor.BuildTimetable(newGroups, file, newLegend)
	return newGroups, newLegend, newTimetable, nil
}

func updateTimetableAndCache(availableGroups *atomic.Value, timetable *atomic.Value, legend *atomic.Value, isInitialRun bool) (hasChanged bool, err error) {
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
		newGroups, newLegend, newTimetable, err := buildNewCache()
		if err != nil {
			return false, err
		}
		availableGroups.Store(newGroups)
		legend.Store(newLegend)
		timetable.Store(newTimetable)
	}

	return hasChanged, nil
}

func main() {
	queries := initializeApp()

	var availableGroups atomic.Value
	var timetable atomic.Value
	var legend atomic.Value

	_, err := updateTimetableAndCache(&availableGroups, &timetable, &legend, true)
	if err != nil {
		slog.Error("Failed to build initial cache", "error", err)
		panic(err)
	}

	ticker := time.NewTicker(time.Duration(global.MINUTES_BETWEEN) * time.Minute)
	done := make(chan bool)

	slog.Info("Starting web server", slog.String("port", global.SERVER_PORT))

	go webServer.Start(queries, &availableGroups, &timetable, &legend)

	slog.Info("Starting timetable update ticker",
		slog.Int("interval_minutes", global.MINUTES_BETWEEN))
	func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				slog.Info("Timetable update triggered", slog.Time("tick_time", t))
				updateAllTimetables(queries, &availableGroups, &timetable, &legend)
			}
		}
	}()
}

func updateAllTimetables(queries db.Querier, availableGroups *atomic.Value, timetable *atomic.Value, legend *atomic.Value) {
	users, err := queries.GetAllUsers(context.Background())
	if err != nil {
		slog.Error("Failed to fetch users", slog.Any("error", err))
		return
	}
	hasChanged, err := updateTimetableAndCache(availableGroups, timetable, legend, false)
	if err != nil {
		slog.Error("Failed to update timetable and cache", slog.Any("error", err))
	}
	allTimetables := timetable.Load().(types.Timetable)

	for _, user := range users {
		if *user.ChosenColumn < 0 || !*user.IsActive {
			slog.Info("Skipping timetable update for inactive user or user without chosen column",
				slog.String("email", *user.Email),
				slog.Int64("column", *user.ChosenColumn))

			continue
		}
		timeSinceSync := time.Since(user.LastSynced)
		if !hasChanged && timeSinceSync.Hours() < float64(global.TIME_BETWEEN_SYNC_HOURS) {
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
		calendar.ImportTimetable(allTimetables, *user.ChosenColumn, utils.GetToken(&user), queries, overrides)
	}
}
