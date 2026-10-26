package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
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

	return h.classesForColumn(*user.ChosenColumn)
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
	return connect.NewResponse(&pkplanerv1.GetClassesResponse{Classes: classes}), nil
}
