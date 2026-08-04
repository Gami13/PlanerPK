package webServer

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"connectrpc.com/connect"
	"golang.org/x/oauth2"

	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
	"github.com/woodpkrs/PlanerPK/gen/pkplaner/v1/pkplanerv1connect"
	"github.com/woodpkrs/PlanerPK/types"

	"github.com/woodpkrs/PlanerPK/calendar"
)

// PKPlanerHandler implements pkplanerv1connect.PKPlanerServiceHandler.
type PKPlanerHandler struct {
	pkplanerv1connect.UnimplementedPKPlanerServiceHandler
	Query db.Querier
	cache *types.Cache
}

func NewPKPlanerHandler(query db.Querier, cache *types.Cache) *PKPlanerHandler {
	if query == nil {
		panic("Query is nil!")
	}
	return &PKPlanerHandler{Query: query, cache: cache}
}

// GetColumns returns available column groups parsed from the Excel header rows.
func (h *PKPlanerHandler) GetColumns(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetColumnsRequest],
) (*connect.Response[pkplanerv1.GetColumnsResponse], error) {
	groups := h.cache.AvailableGroups()

	pbGroups := make([]*pkplanerv1.ColumnGroup, len(*groups))
	for i, g := range *groups {
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
func isValidColumn(column int64, groups []types.ColumnGroup) bool {
	for _, g := range groups {
		if int64(g.Column) == column {
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
	user := ctx.Value(authUserKey).(db.User)

	chosenColumn := int64(req.Msg.ChosenColumn)
	slog.Info("Setting chosen column", slog.Int64("column", chosenColumn))

	groups := h.cache.AvailableGroups()

	if !isValidColumn(chosenColumn, *groups) {
		slog.Warn("Invalid chosen column value", slog.Int64("column", chosenColumn))
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid chosen column value"))
	}

	if err := h.Query.UpdateChosenColumn(ctx, db.UpdateChosenColumnParams{
		ChosenColumn: &chosenColumn,
		AccessToken:  user.AccessToken,
	}); err != nil {
		slog.Error("Failed to update chosen column",
			slog.Int64("column", chosenColumn),
			slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update chosen column: %w", err))
	}

	slog.Info("Triggering timetable import", slog.Int64("column", chosenColumn))
	allTimetables := h.cache.Timetable()
	overrides, err := h.Query.GetOverridesForUser(ctx, user.ID)
	if err != nil {
		slog.Warn("Failed to fetch overrides for initial import, proceeding without", slog.Any("error", err))
		overrides = nil
	}
	var b *types.Buildings
	if h.cache.Buildings() != nil {
		if loaded := h.cache.Buildings(); loaded != nil {
			b = loaded
		}
	}
	go calendar.ImportTimetable(allTimetables, chosenColumn, &oauth2.Token{
		AccessToken:  user.AccessToken,
		TokenType:    user.TokenType,
		RefreshToken: user.RefreshToken,
		Expiry:       user.Expiry,
		ExpiresIn:    user.ExpiresIn,
	}, h.Query, overrides, b)

	return connect.NewResponse(&pkplanerv1.SetChosenColumnResponse{}), nil
}

// SetUserActive sets whether the authenticated account is active.
func (h *PKPlanerHandler) SetUserActive(
	ctx context.Context,
	req *connect.Request[pkplanerv1.SetUserActiveRequest],
) (*connect.Response[pkplanerv1.SetUserActiveResponse], error) {
	user := ctx.Value(authUserKey).(db.User)

	isActive := req.Msg.IsActive
	slog.Info("Setting user active status", slog.Bool("is_active", isActive))

	if err := h.Query.UpdateUserActive(ctx, db.UpdateUserActiveParams{
		AccessToken: user.AccessToken,
		IsActive:    &isActive,
	}); err != nil {
		slog.Error("Failed to update user active status",
			slog.Bool("is_active", isActive),
			slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user active status: %w", err))
	}

	return connect.NewResponse(&pkplanerv1.SetUserActiveResponse{}), nil
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

	user := db.UpsertUserParams{
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
	newUser, err := h.Query.UpsertUser(ctx, user)
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
	user := ctx.Value(authUserKey).(db.User)

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

// startHourRe validates "HH:MM" format.
var startHourRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// GetClasses returns all classes for the authenticated user's chosen column as ClassInfo protos.
func (h *PKPlanerHandler) GetClasses(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetClassesRequest],
) (*connect.Response[pkplanerv1.GetClassesResponse], error) {
	user := ctx.Value(authUserKey).(db.User)

	if user.ChosenColumn == nil || *user.ChosenColumn < 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("no column chosen"))
	}

	allTimetables := h.cache.Timetable()
	timetable, ok := (*allTimetables)[*user.ChosenColumn]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("timetable for chosen column not found"))
	}

	var classes []*pkplanerv1.ClassInfo
	for day, dayClasses := range timetable {
		for _, class := range dayClasses {
			classes = append(classes, &pkplanerv1.ClassInfo{
				Hash:      class.Hash(),
				Name:      class.Name,
				Type:      class.Type,
				Day:       day.String(),
				StartTime: class.StartTime.Format("15:04"),
				EndTime:   class.EndTime.Format("15:04"),
				Location:  class.Location,
				Teacher:   class.Teacher,
			})
		}
	}

	slog.Debug("Returning classes", slog.Int("count", len(classes)))
	return connect.NewResponse(&pkplanerv1.GetClassesResponse{Classes: classes}), nil
}

// GetOverrides returns all class overrides for the authenticated user.
func (h *PKPlanerHandler) GetOverrides(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetOverridesRequest],
) (*connect.Response[pkplanerv1.GetOverridesResponse], error) {
	user := ctx.Value(authUserKey).(db.User)

	rows, err := h.Query.GetOverridesForUser(ctx, user.ID)
	if err != nil {
		slog.Error("Failed to fetch overrides", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch overrides: %w", err))
	}

	ovr := make([]*pkplanerv1.ClassOverride, len(rows))
	for i, row := range rows {
		ovr[i] = &pkplanerv1.ClassOverride{
			ClassHash:     row.ClassHash,
			StartHour:     row.StartHour,
			IsHidden:      row.IsHidden,
			IsTransparent: row.IsTransparent,
		}
	}

	return connect.NewResponse(&pkplanerv1.GetOverridesResponse{Overrides: ovr}), nil
}

// SetOverride creates, updates, or deletes a class override for the authenticated user.
// If all fields are zero-values the override is deleted.
func (h *PKPlanerHandler) SetOverride(
	ctx context.Context,
	req *connect.Request[pkplanerv1.SetOverrideRequest],
) (*connect.Response[pkplanerv1.SetOverrideResponse], error) {
	user := ctx.Value(authUserKey).(db.User)
	ovr := req.Msg.Override

	if ovr == nil || ovr.ClassHash == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("class_hash is required"))
	}

	if ovr.StartHour != "" && !startHourRe.MatchString(ovr.StartHour) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("start_hour must be in HH:MM format"))
	}

	// Delete when all override fields are zero/empty.
	if ovr.StartHour == "" && !ovr.IsHidden && !ovr.IsTransparent {
		slog.Info("Deleting override", slog.String("class_hash", ovr.ClassHash))
		if err := h.Query.DeleteOverride(ctx, db.DeleteOverrideParams{
			UserID:    user.ID,
			ClassHash: ovr.ClassHash,
		}); err != nil {
			slog.Error("Failed to delete override", slog.Any("error", err))
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete override: %w", err))
		}
		return connect.NewResponse(&pkplanerv1.SetOverrideResponse{}), nil
	}

	slog.Info("Upserting override",
		slog.String("class_hash", ovr.ClassHash),
		slog.String("start_hour", ovr.StartHour),
		slog.Bool("is_hidden", ovr.IsHidden),
		slog.Bool("is_transparent", ovr.IsTransparent))

	if err := h.Query.UpsertOverride(ctx, db.UpsertOverrideParams{
		UserID:        user.ID,
		ClassHash:     ovr.ClassHash,
		StartHour:     ovr.StartHour,
		IsHidden:      ovr.IsHidden,
		IsTransparent: ovr.IsTransparent,
	}); err != nil {
		slog.Error("Failed to upsert override", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to upsert override: %w", err))
	}

	return connect.NewResponse(&pkplanerv1.SetOverrideResponse{}), nil
}
