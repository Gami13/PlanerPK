package handler

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
)

// GetOverrides returns all class overrides for the authenticated user.
func (h *PKPlanerHandler) GetOverrides(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetOverridesRequest],
) (*connect.Response[pkplanerv1.GetOverridesResponse], error) {
	user := ctx.Value(AuthUserKey).(db.User)

	rows, err := h.Query.GetOverridesForUser(ctx, user.ID)
	if err != nil {
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
	user := ctx.Value(AuthUserKey).(db.User)
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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to upsert override: %w", err))
	}

	return connect.NewResponse(&pkplanerv1.SetOverrideResponse{}), nil
}
