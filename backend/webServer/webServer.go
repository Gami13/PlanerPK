package webServer

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/gen/pkplaner/v1/pkplanerv1connect"
	"github.com/woodpkrs/PlanerPK/global"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/webServer/handler"

	"connectrpc.com/connect"

	_ "modernc.org/sqlite"
)

func corsMiddleware(next http.Handler) http.Handler {
	origins := global.Config.AllowedOrigins

	slog.Info("CORS middleware initialized", slog.String("allowed_origins", strings.Join(origins, ", ")))

	// Build a lookup table so we can validate request origins quickly per request.
	allowedOrigins := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowedOrigins[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if _, ok := allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version, Connect-Timeout-Ms, Grpc-Timeout, X-Grpc-Web, X-User-Agent")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Start(db db.Querier, cache *types.Cache) {
	query := db
	pkPlanerHandler := handler.NewPKPlanerHandler(query, cache)

	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(LoggingInterceptor(), AuthInterceptor(query))
	rpcPath, rpcHandler := pkplanerv1connect.NewPKPlanerServiceHandler(pkPlanerHandler, interceptors)

	mux.Handle(rpcPath, rpcHandler)

	mux.Handle("/", http.HandlerFunc(ServeFrontend))

	corsHandler := corsMiddleware(mux)
	server := &http.Server{
		Addr:      ":" + global.ServerPort,
		Handler:   corsHandler,
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
