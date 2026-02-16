package webServer

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"

	pkplanerv1 "github.com/WoodPKrs/GoogleCalendarTimetable/gen/v1"
	"github.com/WoodPKrs/GoogleCalendarTimetable/gen/v1/pkplanerv1connect"

	"github.com/WoodPKrs/GoogleCalendarTimetable/calendar"
	"github.com/WoodPKrs/GoogleCalendarTimetable/database"
	"github.com/WoodPKrs/GoogleCalendarTimetable/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
)

// PKPlanerHandler implements pkplanerv1connect.PKPlanerServiceHandler.
type PKPlanerHandler struct {
	pkplanerv1connect.UnimplementedPKPlanerServiceHandler
	Query *database.Queries
}

func NewPKPlanerHandler(query *database.Queries) *PKPlanerHandler {
	if query == nil {
		panic("Query is nil!")
	}
	return &PKPlanerHandler{Query: query}
}

// GetColumns returns available column groups parsed from the Excel header rows.
func (h *PKPlanerHandler) GetColumns(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetColumnsRequest],
) (*connect.Response[pkplanerv1.GetColumnsResponse], error) {
	groups, err := extractor.GetAvailableGroups()
	if err != nil {
		slog.Error("Failed to get available groups", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get available groups: %w", err))
	}

	pbGroups := make([]*pkplanerv1.ColumnGroup, len(groups))
	for i, g := range groups {
		pbGroups[i] = &pkplanerv1.ColumnGroup{
			Column:   int32(g.Column),
			Label:    g.Label,
			Category: g.Category,
		}
	}

	return connect.NewResponse(&pkplanerv1.GetColumnsResponse{
		Columns: pbGroups,
	}), nil
}

// isValidColumn checks whether the given column index is present in the available groups.
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

// SetChosenColumn sets the chosen column for the authenticated user.
func (h *PKPlanerHandler) SetChosenColumn(
	ctx context.Context,
	req *connect.Request[pkplanerv1.SetChosenColumnRequest],
) (*connect.Response[pkplanerv1.SetChosenColumnResponse], error) {
	accessToken := req.Msg.AccessToken
	if accessToken == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing access token"))
	}

	chosenColumn := int(req.Msg.ChosenColumn)
	slog.Info("Setting chosen column", slog.Int("column", chosenColumn))

	if !isValidColumn(chosenColumn) {
		slog.Warn("Invalid chosen column value", slog.Int("column", chosenColumn))
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid chosen column value"))
	}

	if err := h.Query.UpdateChosenColumn(accessToken, int64(chosenColumn)); err != nil {
		slog.Error("Failed to update chosen column",
			slog.Int("column", chosenColumn),
			slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update chosen column: %w", err))
	}

	token, err := h.Query.GetTokenByAccessToken(accessToken)
	if err != nil {
		slog.Error("Failed to get token", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get token: %w", err))
	}

	slog.Info("Triggering timetable import", slog.Int("column", chosenColumn))
	go calendar.ImportTimetable(int64(chosenColumn), token)

	return connect.NewResponse(&pkplanerv1.SetChosenColumnResponse{}), nil
}

// Authorize exchanges an OAuth code for user credentials and returns user info.
func (h *PKPlanerHandler) Authorize(
	ctx context.Context,
	req *connect.Request[pkplanerv1.AuthorizeRequest],
) (*connect.Response[pkplanerv1.AuthorizeResponse], error) {
	code := req.Msg.Code
	if code == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing code parameter"))
	}

	slog.Info("Processing authorization", slog.String("endpoint", "Authorize"))

	tok, err := calendar.AuthorizeFromCode(code)
	if err != nil {
		slog.Error("Authorization failed", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization failed: %w", err))
	}

	userInfo, err := calendar.GetUserInfo(tok)
	if err != nil {
		slog.Error("Failed to get user info", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user info: %w", err))
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
	newUser, err := h.Query.UpsertUser(&user)
	if err != nil {
		slog.Error("Failed to save user",
			slog.String("email", userInfo.Email),
			slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to save user: %w", err))
	}

	slog.Info("User authorized successfully", slog.String("email", *newUser.Email))

	return connect.NewResponse(&pkplanerv1.AuthorizeResponse{
		AccessToken:  newUser.AccessToken,
		Email:        *newUser.Email,
		ChosenColumn: *newUser.ChosenColumn,
		Name:         *newUser.Name,
		Picture:      *newUser.Picture,
	}), nil
}

// GetAuthURL generates a Google OAuth authorization URL.
func (h *PKPlanerHandler) GetAuthURL(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetAuthURLRequest],
) (*connect.Response[pkplanerv1.GetAuthURLResponse], error) {
	slog.Debug("Generating auth URL", slog.String("endpoint", "GetAuthURL"))
	authURL, err := calendar.GenerateAuthURL()
	if err != nil {
		slog.Error("Failed to generate auth URL", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate auth URL: %w", err))
	}

	return connect.NewResponse(&pkplanerv1.GetAuthURLResponse{
		Url: authURL,
	}), nil
}

// GetUserInfo returns information about the authenticated user.
func (h *PKPlanerHandler) GetUserInfo(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetUserInfoRequest],
) (*connect.Response[pkplanerv1.GetUserInfoResponse], error) {
	accessToken := req.Msg.AccessToken
	if accessToken == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing access token"))
	}

	user, err := h.Query.GetUserByAccessToken(accessToken)
	if err != nil {
		slog.Warn("Invalid access token or user not found", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("unauthorized: invalid or expired access token"))
	}

	slog.Debug("Retrieved user info", slog.String("email", *user.Email))

	return connect.NewResponse(&pkplanerv1.GetUserInfoResponse{
		Email:        *user.Email,
		Name:         *user.Name,
		GivenName:    *user.GivenName,
		Picture:      *user.Picture,
		IsActive:     *user.IsActive,
		ChosenColumn: *user.ChosenColumn,
	}), nil
}
