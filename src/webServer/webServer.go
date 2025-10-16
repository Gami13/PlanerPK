package webServer

import (
	"GoogleCalendarTimetable/src/database"
	"context"
	"database/sql"
	"net/http"
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
	//TODO: @Frytak285 Pls change so its not in memory, this is a placeholder
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	query := database.New(db)
	hctx := database.NewEndpointContext(query, ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /authorize", hctx.AuthorizeEndpoint)

	handler := corsMiddleware(mux)
	http.ListenAndServe(":8080", handler)
}
