package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/calendar"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
	"github.com/woodpkrs/PlanerPK/types"
)

// startHourRe validates "HH:MM" format.
var startHourRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// GetClasses returns all classes for the authenticated user's chosen column as ClassInfo protos.
func (h *PKPlanerHandler) GetClasses(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetClassesRequest],
) (*connect.Response[pkplanerv1.GetClassesResponse], error) {
	user := ctx.Value(AuthUserKey).(db.User)

	if user.ChosenColumn == nil || *user.ChosenColumn < 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("no column chosen"))
	}

	rows, err := h.Query.GetOverridesForUser(ctx, user.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch overrides: %w", err))
	}
	all := h.cache.Timetable()
	if _, ok := (*all)[*user.ChosenColumn]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("timetable not found"))
	}
	return classesResponse(calendar.UserTimetable(*all, *user.ChosenColumn, rows, *h.cache.AvailableGroups())), nil
}

// GetTimetable exposes only the published schedule, never user overrides.
func (h *PKPlanerHandler) GetTimetable(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetTimetableRequest],
) (*connect.Response[pkplanerv1.GetClassesResponse], error) {
	column := int64(req.Msg.Column)
	if !isValidColumn(column, *h.cache.AvailableGroups()) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("timetable not found"))
	}
	return h.classesForColumn(column)
}

func (h *PKPlanerHandler) classesForColumn(column int64) (*connect.Response[pkplanerv1.GetClassesResponse], error) {
	allTimetables := h.cache.Timetable()
	timetable, ok := (*allTimetables)[column]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("timetable for chosen column not found"))
	}

	return classesResponse(timetable), nil
}

func classesResponse(timetable types.GroupTimetable) *connect.Response[pkplanerv1.GetClassesResponse] {
	var classes []*pkplanerv1.ClassInfo
	for day, dayClasses := range timetable {
		for _, class := range dayClasses {
			classes = append(classes, &pkplanerv1.ClassInfo{
				Hash:       class.Hash(),
				Name:       class.Name,
				Type:       class.Type,
				Day:        day.String(),
				StartTime:  class.StartTime.Format("15:04"),
				EndTime:    class.EndTime.Format("15:04"),
				Location:   class.Location,
				Teacher:    class.Teacher,
				Audience:   class.Audience,
				Summary:    class.Summary(),
				Recurrence: class.AdditionalInfo,
			})
		}
	}

	slog.Debug("Returning classes", slog.Int("count", len(classes)))
	return connect.NewResponse(&pkplanerv1.GetClassesResponse{Classes: classes})
}

func (h *PKPlanerHandler) GetImportGroups(ctx context.Context, req *connect.Request[pkplanerv1.GetColumnsRequest]) (*connect.Response[pkplanerv1.GetColumnsResponse], error) {
	user := ctx.Value(AuthUserKey).(db.User)
	if user.ChosenColumn == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("choose a group first"))
	}
	groups := *h.cache.AvailableGroups()
	response := &pkplanerv1.GetColumnsResponse{}
	for _, group := range groups {
		if types.CanImportGroup(groups, *user.ChosenColumn, int64(group.Column)) {
			response.Columns = append(response.Columns, &pkplanerv1.ColumnGroup{Column: int32(group.Column), Label: group.Label, Category: group.Category})
		}
	}
	return connect.NewResponse(response), nil
}
