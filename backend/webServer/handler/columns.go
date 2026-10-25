package handler

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/calendar"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/utils"
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
	h.syncUserTimetable(user)
	return connect.NewResponse(&pkplanerv1.SetChosenColumnResponse{}), nil
}

// User setting changes sync immediately, independently of the global cache gate.
func (h *PKPlanerHandler) syncUserTimetable(user db.User) {
	allTimetables := h.cache.Timetable()
	state := h.cache.Snapshot()
	b := h.cache.Buildings()
	token := utils.GetToken(&user)
	go func() {
		if err := calendar.ImportTimetable(allTimetables, token, h.Query, state, b); err != nil {
			slog.Error("Timetable sync failed", slog.Int64("user_id", user.ID), slog.Any("error", err))
		}
	}()
}
