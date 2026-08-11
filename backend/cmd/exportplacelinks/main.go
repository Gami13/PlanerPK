package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

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

	global.LoadEnvVariables()

	outPath := *outPathFlag
	if outPath == "" {
		outPath = filepath.Join(global.StateDir(), "place_links.txt")
	}

	database, err := sql.Open(global.DatabaseDriver(), global.DatabaseSource())
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

	// Query normalized classrooms and their associated building-place details.
	rows, err := database.QueryContext(ctx, `
		SELECT cb.classroom_key, cb.building_key, COALESCE(bp.place_id, ''),
			COALESCE(bp.display_name, ''), COALESCE(bp.address, '')
		FROM classroom_buildings cb
		JOIN building_places bp ON bp.building_key = cb.building_key
		ORDER BY cb.classroom_key ASC
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying classrooms: %v\n", err)
		os.Exit(1)
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error in rows: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var classroomPlaces []PlaceInfo
	for rows.Next() {
		var cName, bKey, placeID, bName, bAddr string
		if err := rows.Scan(&cName, &bKey, &placeID, &bName, &bAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
			continue
		}

		info := buildPlaceInfo(cName, bKey, placeID, bName, bAddr)
		classroomPlaces = append(classroomPlaces, info)
	}

	var buildingPlaces []PlaceInfo
	for _, b := range allBuildings {
		placeID := ""
		if b.PlaceID != nil {
			placeID = *b.PlaceID
		}
		name := ""
		if b.DisplayName != nil {
			name = *b.DisplayName
		}
		addr := ""
		if b.Address != nil {
			addr = *b.Address
		}

		info := buildPlaceInfo("", b.BuildingKey, placeID, name, addr)
		buildingPlaces = append(buildingPlaces, info)
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
	fmt.Fprintf(outFile, "                         DATABASE PLACE HYPERLINKS                              \n")
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

func buildPlaceInfo(classroomName, bKey, placeID, bName, bAddr string) PlaceInfo {
	info := PlaceInfo{
		ClassroomName: classroomName,
		BuildingKey:   bKey,
		BuildingName:  bName,
		Address:       bAddr,
		GooglePlaceID: placeID,
	}

	if info.GooglePlaceID != "" {
		label := "Google Maps"
		if info.BuildingName != "" && info.Address != "" {
			label = fmt.Sprintf("%s, %s", info.BuildingName, info.Address)
		} else if info.BuildingName != "" {
			label = info.BuildingName
		} else if info.Address != "" {
			label = info.Address
		}
		info.Label = label
		info.MapsURL = fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=Politechnika+Krakowska&query_place_id=%s", info.GooglePlaceID)
		info.HTMLAnchor = fmt.Sprintf("<a href=\"%s\">%s</a>", info.MapsURL, info.Label)
	} else if info.BuildingName != "" || info.Address != "" {
		if info.BuildingName != "" && info.Address != "" {
			info.Label = fmt.Sprintf("%s, %s", info.BuildingName, info.Address)
		} else if info.BuildingName != "" {
			info.Label = info.BuildingName
		} else {
			info.Label = info.Address
		}
	}

	return info
}
