package webServer

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/gen/db"
	"github.com/WoodPKrs/GoogleCalendarTimetable/gen/pkplaner/v1/pkplanerv1connect"
	"github.com/WoodPKrs/GoogleCalendarTimetable/global"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

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

func Start(db db.Querier) {
	query := db
	handler := NewPKPlanerHandler(query)

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

	// Mount ConnectRPC service behind rate limiting.
	interceptors := connect.WithInterceptors(AuthInterceptor(query))
	rpcPath, rpcHandler := pkplanerv1connect.NewPKPlanerServiceHandler(handler, interceptors)

	// Wrap the SetChosenColumn endpoint with the strict policy.
	mux.Handle(rpcPath, rateLimitMiddleware(rateLimiter, generalPolicy, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/SetChosenColumn") {
			rateLimitMiddleware(rateLimiter, setChosenColumnPolicy, rpcHandler).ServeHTTP(w, r)
			return
		}
		rpcHandler.ServeHTTP(w, r)
	})))

	// Frontend static files (catch-all).

	mux.Handle("/", rateLimitMiddleware(rateLimiter, generalPolicy, http.HandlerFunc(ServeFrontend)))

	// CORS + h2c for HTTP/2 cleartext (Connect protocol).
	corsHandler := corsMiddleware(mux)
	h2cHandler := h2c.NewHandler(corsHandler, &http2.Server{})

	slog.Info("Web server listening", slog.String("port", global.SERVER_PORT))
	if err := http.ListenAndServe(":"+global.SERVER_PORT, h2cHandler); err != nil {
		slog.Error("Web server failed", slog.Any("error", err))
	}
}
