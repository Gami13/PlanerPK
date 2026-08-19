package handler

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/calendar"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
	"github.com/woodpkrs/PlanerPK/types"
	"golang.org/x/oauth2"
)

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
	user := ctx.Value(AuthUserKey).(db.User)

	chosenColumn := int64(req.Msg.ChosenColumn)
	slog.Info("Setting chosen column", slog.Int64("column", chosenColumn))

	groups := h.cache.AvailableGroups()

	if !isValidColumn(chosenColumn, *groups) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid chosen column value"))
	}

	if err := h.Query.UpdateChosenColumn(ctx, db.UpdateChosenColumnParams{
		ChosenColumn: &chosenColumn,
		AccessToken:  user.AccessToken,
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update chosen column: %w", err))
	}

	slog.Info("Triggering timetable import", slog.Int64("column", chosenColumn))
	allTimetables := h.cache.Timetable()
	overrides, err := h.Query.GetOverridesForUser(ctx, user.ID)
	if err != nil {
		slog.Warn("Failed to fetch overrides for initial import, proceeding without", slog.Any("error", err))
		overrides = nil
	}
	var b *buildings.Buildings
	if loaded := h.cache.Buildings(); loaded != nil {
		b = loaded
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
