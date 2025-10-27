package webServer

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/database"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/global"

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

func Start(db *sql.DB) {
	ctx := context.Background()

	query := database.New(db)
	hctx := database.NewEndpointContext(query, ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/authorize", hctx.AuthorizeEndpoint)
	mux.HandleFunc("PUT /api/setChosenColumn", hctx.SetChosenColumnEndpoint)
	mux.HandleFunc("GET /api/getAuthURL", hctx.GetAuthURLEndpoint)
	mux.HandleFunc("GET /api/getUserInfo", hctx.GetUserInfoEndpoint)

	handler := corsMiddleware(mux)

	slog.Info("Web server listening", slog.String("port", global.SERVER_PORT))
	if err := http.ListenAndServe(":"+global.SERVER_PORT, handler); err != nil {
		slog.Error("Web server failed", slog.Any("error", err))
	}
}

