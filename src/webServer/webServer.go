package webServer

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

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
	hctx := NewEndpointContext(query, ctx)

	mux := http.NewServeMux()
	rateLimiter := NewRateLimiter()
	generalPolicy := RateLimitPolicy{
		Name:     "general",
		Requests: global.GENERAL_RATE_LIMIT_PER_MINUTE,
		Window:   time.Minute,
	}
	setChosenColumnPolicy := RateLimitPolicy{
		Name:        "setChosenColumn-strict",
		Requests:    3,
		Window:      time.Hour,
		BanDuration: 7 * 24 * time.Hour,
	}

	registerGeneral := func(pattern string, handler http.HandlerFunc) {
		mux.Handle(pattern, rateLimitMiddleware(rateLimiter, generalPolicy, handler))
	}

	registerGeneral("GET /api/authorize", hctx.AuthorizeEndpoint)
	registerGeneral("GET /api/getAuthURL", hctx.GetAuthURLEndpoint)
	registerGeneral("GET /api/getUserInfo", hctx.GetUserInfoEndpoint)
	mux.Handle("PUT /api/setChosenColumn", rateLimitMiddleware(rateLimiter, generalPolicy, rateLimitMiddleware(rateLimiter, setChosenColumnPolicy, http.HandlerFunc(hctx.SetChosenColumnEndpoint))))
	//serves the frontend
	mux.Handle("/", http.FileServer(http.Dir("./static/")))

	handler := corsMiddleware(mux)

	slog.Info("Web server listening", slog.String("port", global.SERVER_PORT))
	if err := http.ListenAndServe(":"+global.SERVER_PORT, handler); err != nil {
		slog.Error("Web server failed", slog.Any("error", err))
	}
}
