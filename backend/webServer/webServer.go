package webServer

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/gen/pkplaner/v1/pkplanerv1connect"
	"github.com/woodpkrs/PlanerPK/global"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/webServer/handler"

	"connectrpc.com/connect"

	_ "modernc.org/sqlite"
)

type Credentials struct {
	Web struct {
		ClientID                string   `json:"client_id"`
		ProjectID               string   `json:"project_id"`
		AuthURI                 string   `json:"auth_uri"`
		TokenURI                string   `json:"token_uri"`
		AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
		ClientSecret            string   `json:"client_secret"`
		RedirectURIs            []string `json:"redirect_uris"`
		JavascriptOrigins       []string `json:"javascript_origins"`
	} `json:"web"`
}

func readAllowedOrigins() ([]string, error) {
	b, err := os.ReadFile(global.CALENDAR_CREDENTIALS_PATH)
	if err != nil {
		slog.Error("Failed to read credentials file", slog.Any("error", err))
		return nil, err
	}
	var creds Credentials
	err = json.Unmarshal(b, &creds)
	if err != nil {
		slog.Error("Failed to parse credentials file", slog.Any("error", err))
		return nil, err
	}
	return creds.Web.JavascriptOrigins, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	origins, err := readAllowedOrigins()
	if err != nil {
		slog.Error("Failed to read allowed origins for CORS middleware", slog.Any("error", err))
		panic(err)
	}

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

	interceptors := connect.WithInterceptors(AuthInterceptor(query))
	rpcPath, rpcHandler := pkplanerv1connect.NewPKPlanerServiceHandler(pkPlanerHandler, interceptors)

	mux.Handle(rpcPath, rpcHandler)

	mux.Handle("/", http.HandlerFunc(ServeFrontend))

	corsHandler := corsMiddleware(mux)
	server := &http.Server{
		Addr:      ":" + global.SERVER_PORT,
		Handler:   corsHandler,
		Protocols: &http.Protocols{},
	}
	server.Protocols.SetHTTP1(true)
	server.Protocols.SetHTTP2(true)
	server.Protocols.SetUnencryptedHTTP2(true)

	slog.Info("Web server listening", slog.String("port", global.SERVER_PORT))
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Web server failed", slog.Any("error", err))
	}
}
