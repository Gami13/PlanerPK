package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	_ "modernc.org/sqlite"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/extractor"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/global"
)

type PlaceInfo struct {
	ClassroomName string
	BuildingKey   string
	BuildingName  string
	Address       string
	GooglePlaceID string
	Label         string
	MapsURL       string
	HTMLAnchor    string
}

func main() {
	outPathFlag := flag.String("out", "", "Output file path (default: <STATE_DIR>/place_links.txt)")
	flag.Parse()

	global.Load()
	config := global.Config

	outPath := *outPathFlag
	if outPath == "" {
		outPath = filepath.Join(config.StateDir, "place_links.txt")
	}

	database, err := sql.Open(global.DatabaseDriver, config.DatabaseSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	queries := db.New(database)
	ctx := context.Background()

	allBuildings, err := queries.GetBuildingPlaces(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching building places from DB: %v\n", err)
		os.Exit(1)
	}

	file, err := extractor.OpenCurrentTimetableFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening current timetable: %v\n", err)
		os.Exit(1)
	}
	defer extractor.CloseFile(file)

	legendBuildings := buildings.BuildBuildingsFromLegend(extractor.ExtractLegend(file))
	placesByBuildingKey := make(map[string]db.BuildingPlace, len(allBuildings))
	for _, building := range allBuildings {
		placesByBuildingKey[building.BuildingKey] = building
	}

	var classroomPlaces []PlaceInfo
	classroomKeys := slices.Sorted(maps.Keys(legendBuildings.BuildingKeyByClassroomKey))
	for _, classroomKey := range classroomKeys {
		buildingKey := legendBuildings.BuildingKeyByClassroomKey[classroomKey]
		classroomPlaces = append(classroomPlaces, buildPlaceInfo(classroomKey, buildingKey, placesByBuildingKey[buildingKey]))
	}

	var buildingPlaces []PlaceInfo
	for _, b := range allBuildings {
		buildingPlaces = append(buildingPlaces, buildPlaceInfo("", b.BuildingKey, b))
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	fmt.Fprintf(outFile, "================================================================================\n")
	fmt.Fprintf(outFile, "                          LEGEND PLACE HYPERLINKS                               \n")
	fmt.Fprintf(outFile, "================================================================================\n\n")

	fmt.Fprintf(outFile, "--------------------------------------------------------------------------------\n")
	fmt.Fprintf(outFile, " 1. UNIQUE CLASSROOMS AND THEIR ATTACHED HYPERLINKS (%d found)\n", len(classroomPlaces))
	fmt.Fprintf(outFile, "--------------------------------------------------------------------------------\n\n")

	for i, p := range classroomPlaces {
		fmt.Fprintf(outFile, "[Classroom #%d] %s\n", i+1, p.ClassroomName)
		if p.BuildingKey != "" {
			fmt.Fprintf(outFile, "  Building Key    : %s\n", p.BuildingKey)
			fmt.Fprintf(outFile, "  Google Place ID : %s\n", p.GooglePlaceID)
			fmt.Fprintf(outFile, "  Label           : %s\n", p.Label)
			fmt.Fprintf(outFile, "  Maps URL        : %s\n", p.MapsURL)
			fmt.Fprintf(outFile, "  HTML Hyperlink  : %s\n", p.HTMLAnchor)
		} else {
			fmt.Fprintf(outFile, "  Status          : Unmapped (no building linked)\n")
		}
		fmt.Fprintf(outFile, "\n")
	}

	fmt.Fprintf(outFile, "--------------------------------------------------------------------------------\n")
	fmt.Fprintf(outFile, " 2. UNIQUE BUILDINGS IN DATABASE (%d found)\n", len(buildingPlaces))
	fmt.Fprintf(outFile, "--------------------------------------------------------------------------------\n\n")

	for i, b := range buildingPlaces {
		fmt.Fprintf(outFile, "[Building #%d] %s (Key: %s)\n", i+1, b.Label, b.BuildingKey)
		fmt.Fprintf(outFile, "  Google Place ID : %s\n", b.GooglePlaceID)
		fmt.Fprintf(outFile, "  Address         : %s\n", b.Address)
		fmt.Fprintf(outFile, "  Maps URL        : %s\n", b.MapsURL)
		fmt.Fprintf(outFile, "  HTML Hyperlink  : %s\n", b.HTMLAnchor)
		fmt.Fprintf(outFile, "\n")
	}

	fmt.Printf("Successfully exported hyperlinks for %d unique classrooms and %d unique buildings to %s\n",
		len(classroomPlaces), len(buildingPlaces), outPath)
}

func buildPlaceInfo(classroomName, buildingKey string, place db.BuildingPlace) PlaceInfo {
	building := buildings.BuildingFromPlace(place)
	info := PlaceInfo{
		ClassroomName: classroomName,
		BuildingKey:   buildingKey,
		BuildingName:  building.Name,
		Address:       building.Address,
		GooglePlaceID: building.GooglePlaceID,
		Label:         building.MapLabel(),
		MapsURL:       building.MapURL(),
		HTMLAnchor:    building.MapHTMLAnchor(),
	}
	return info
}
