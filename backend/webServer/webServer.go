package webServer

import (
	"log/slog"
	"net/http"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/gen/pkplaner/v1/pkplanerv1connect"
	"github.com/woodpkrs/PlanerPK/global"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/webServer/handler"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"

	_ "modernc.org/sqlite"
)

func corsMiddleware(next http.Handler) http.Handler {
	if len(global.Config.AllowedOrigins) == 0 {
		return next
	}
	return cors.New(cors.Options{
		AllowedOrigins: global.Config.AllowedOrigins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: append(connectcors.AllowedHeaders(), "Authorization"),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         7200,
	}).Handler(next)
}

func Start(db db.Querier, cache *types.Cache) {
	query := db
	pkPlanerHandler := handler.NewPKPlanerHandler(query, cache)

	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(LoggingInterceptor(), AuthInterceptor(query))
	rpcPath, rpcHandler := pkplanerv1connect.NewPKPlanerServiceHandler(pkPlanerHandler, interceptors)

	mux.Handle(rpcPath, corsMiddleware(rpcHandler))

	mux.Handle("/", http.HandlerFunc(ServeFrontend))

	server := &http.Server{
		Addr:      ":" + global.ServerPort,
		Handler:   mux,
		Protocols: &http.Protocols{},
	}
	server.Protocols.SetHTTP1(true)
	server.Protocols.SetHTTP2(true)
	server.Protocols.SetUnencryptedHTTP2(true)

	slog.Info("Web server listening", slog.String("port", global.ServerPort))
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Web server failed", slog.Any("error", err))
	}
}
