package webServer

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/database"

	_ "modernc.org/sqlite"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Start() {

	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:database.db?cache=shared&mode=rwc&_busy_timeout=5000")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	query := database.New(db)
	hctx := database.NewEndpointContext(query, ctx)

	err = database.InitializeDatabase(db)
	if err != nil {
		panic(err)
	}
	fmt.Println("Initialized database successfully.")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /authorize", hctx.AuthorizeEndpoint)
	mux.HandleFunc("PUT /setChosenColumn", hctx.SetChosenColumnEndpoint)

	handler := corsMiddleware(mux)
	http.ListenAndServe(":8080", handler)
}
