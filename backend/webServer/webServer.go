package webServer

import (
	"log/slog"
	"net/http"

	"github.com/woodpkrs/PlanerPK/cache"
	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/gen/pkplaner/v1/pkplanerv1connect"
	"github.com/woodpkrs/PlanerPK/utils"
	"github.com/woodpkrs/PlanerPK/webServer/handler"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"

	_ "modernc.org/sqlite"
)

func corsMiddleware(next http.Handler) http.Handler {
	if len(config.Config.AllowedOrigins) == 0 {
		return next
	}
	return cors.New(cors.Options{
		AllowedOrigins: config.Config.AllowedOrigins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: append(connectcors.AllowedHeaders(), "Authorization"),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         7200,
	}).Handler(next)
}

func Start(db db.Querier, cache *cache.Cache) {
	query := db
	pkPlanerHandler := handler.NewPKPlanerHandler(query, cache)

	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(LoggingInterceptor(), AuthInterceptor(query))
	rpcPath, rpcHandler := pkplanerv1connect.NewPKPlanerServiceHandler(pkPlanerHandler, interceptors)

	mux.Handle(rpcPath, corsMiddleware(rpcHandler))

	mux.Handle("/", http.HandlerFunc(ServeFrontend))

	server := &http.Server{
		Addr:      ":" + config.ServerPort,
		Handler:   mux,
		Protocols: &http.Protocols{},
	}
	server.Protocols.SetHTTP1(true)
	server.Protocols.SetHTTP2(true)
	server.Protocols.SetUnencryptedHTTP2(true)

	slog.Info("Web server listening", slog.String("port", config.ServerPort))
	if err := server.ListenAndServe(); err != nil {
		utils.NotifyError("Web server failed", slog.Any("error", err))
	}
}
