package webServer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/WoodPKrs/GoogleCalendarTimetable/calendar"
	"github.com/WoodPKrs/GoogleCalendarTimetable/database"
	"github.com/WoodPKrs/GoogleCalendarTimetable/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
)

type EndpointContext struct {
	Query *database.Queries
	Ctx   context.Context
}

func NewEndpointContext(query *database.Queries, ctx context.Context) *EndpointContext {
	if query == nil {
		panic("Query is nil!")
	}
	if ctx == nil {
		panic("Context is nil!")
	}

	return &EndpointContext{Query: query, Ctx: ctx}
}

// GET /api/columns
// Returns available column groups parsed from the Excel header rows. No auth required.
func (ec *EndpointContext) GetColumnsEndpoint(w http.ResponseWriter, r *http.Request) {
	groups, err := extractor.GetAvailableGroups()
	if err != nil {
		slog.Error("Failed to get available groups", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to get available groups: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		slog.Error("Failed to encode columns response", slog.Any("error", err))
	}
}

// isValidColumn checks whether the given column index is present in the available groups.
// Falls back to a simple range check if parsing fails.
func isValidColumn(column int) bool {
	groups, err := extractor.GetAvailableGroups()
	if err != nil {
		slog.Warn("Column validation falling back to range check", slog.Any("error", err))
		return column >= 0 && column <= global.LAST_COLUMN
	}
	for _, g := range groups {
		if g.Column == column {
			return true
		}
	}
	return false
}

// PUT /setChosenColumn
// in body: chosenColumn: int
// sets the chosen column for the user associated with the access token in the Authorization header
func (ec *EndpointContext) SetChosenColumnEndpoint(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		slog.Warn("Missing Authorization header", slog.String("endpoint", "SetChosenColumn"))
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	var requestBody struct {
		ChosenColumn int `json:"chosenColumn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		slog.Error("Failed to parse request body",
			slog.String("endpoint", "SetChosenColumn"),
			slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return
	}

	slog.Info("Setting chosen column", slog.Int("column", requestBody.ChosenColumn))

	if !isValidColumn(requestBody.ChosenColumn) {
		slog.Warn("Invalid chosen column value", slog.Int("column", requestBody.ChosenColumn))
		http.Error(w, "Invalid chosen column value", http.StatusBadRequest)
		return
	}

	if err := ec.Query.UpdateChosenColumn(accessToken, int64(requestBody.ChosenColumn)); err != nil {
		slog.Error("Failed to update chosen column",
			slog.Int("column", requestBody.ChosenColumn),
			slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to update chosen column: %v", err), http.StatusInternalServerError)
		return
	}

	// Update the timetable
	token, err := ec.Query.GetTokenByAccessToken(accessToken)
	if err != nil {
		slog.Error("Failed to get token", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	slog.Info("Triggering timetable import", slog.Int("column", requestBody.ChosenColumn))
	go calendar.ImportTimetable(int64(requestBody.ChosenColumn), token)
	w.WriteHeader(http.StatusOK)
}

func (ec *EndpointContext) AuthorizeEndpoint(w http.ResponseWriter, r *http.Request) {
	// Get the oauth code from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		slog.Warn("Missing code parameter", slog.String("endpoint", "Authorize"))
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	slog.Info("Processing authorization", slog.String("endpoint", "Authorize"))

	tok, err := calendar.AuthorizeFromCode(code)
	if err != nil {
		slog.Error("Authorization failed", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Authorization failed: %v", err), http.StatusInternalServerError)
		return
	}

	userInfo, err := calendar.GetUserInfo(tok)
	if err != nil {
		slog.Error("Failed to get user info", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}
	user := database.User{
		Email:        &userInfo.Email,
		Name:         &userInfo.Name,
		GivenName:    &userInfo.GivenName,
		Picture:      &userInfo.Picture,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
		ExpiresIn:    tok.ExpiresIn,
	}
	newUser, err := ec.Query.UpsertUser(&user)
	if err != nil {
		slog.Error("Failed to save user",
			slog.String("email", userInfo.Email),
			slog.Any("error", err))
		return
	}

	slog.Info("User authorized successfully",
		slog.String("email", *newUser.Email))

	response := database.AuthorizeResponse{
		AccessToken:  newUser.AccessToken,
		Email:        *newUser.Email,
		ChosenColumn: *newUser.ChosenColumn,
		Name:         *newUser.Name,
		Picture:      *newUser.Picture,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		slog.Error("Failed to marshal response", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to marshal response: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(responseBytes); err != nil {
		slog.Error("Failed to write response", slog.Any("error", err))
	}
}

func (ec *EndpointContext) GetAuthURLEndpoint(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Generating auth URL", slog.String("endpoint", "GetAuthURL"))
	authURL, err := calendar.GenerateAuthURL()
	if err != nil {
		slog.Error("Failed to generate auth URL", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to generate auth URL: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte(authURL)); err != nil {
		slog.Error("Failed to write response", slog.Any("error", err))
	}
}

func (ec *EndpointContext) GetUserInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		slog.Warn("Missing Authorization header", slog.String("endpoint", "GetUserInfo"))
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	user, err := ec.Query.GetUserByAccessToken(accessToken)
	if err != nil {
		slog.Warn("Invalid access token or user not found", slog.Any("error", err))
		http.Error(w, "Unauthorized: invalid or expired access token", http.StatusUnauthorized)
		return
	}

	slog.Debug("Retrieved user info", slog.String("email", *user.Email))

	response := database.UserInfoResponse{
		Email:        *user.Email,
		Name:         *user.Name,
		GivenName:    *user.GivenName,
		Picture:      *user.Picture,
		IsActive:     *user.IsActive,
		ChosenColumn: *user.ChosenColumn,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		slog.Error("Failed to marshal response", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to marshal response: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(responseBytes); err != nil {
		slog.Error("Failed to write response", slog.Any("error", err))
	}
}

var fs = http.FileServer(http.Dir("./static/"))

func (ec *EndpointContext) ServeFrontend(w http.ResponseWriter, r *http.Request) {
	// Prevent caching of env.js since it's dynamically generated at runtime
	if r.URL.Path == "/env.js" {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}

	if r.URL.Path != "/" {
		fullPath := "./static/" + strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		_, err := os.Stat(fullPath)
		if err != nil {
			if !os.IsNotExist(err) {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			// Requested file does not exist so we return the default (resolves to index.html)
			r.URL.Path = "/"
		}
	}
	fs.ServeHTTP(w, r)
}
