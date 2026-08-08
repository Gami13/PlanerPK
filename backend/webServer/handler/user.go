package handler

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
)

// SetUserActive sets whether the authenticated account is active.
func (h *PKPlanerHandler) SetUserActive(
	ctx context.Context,
	req *connect.Request[pkplanerv1.SetUserActiveRequest],
) (*connect.Response[pkplanerv1.SetUserActiveResponse], error) {
	user := ctx.Value(AuthUserKey).(db.User)

	isActive := req.Msg.IsActive
	slog.Info("Setting user active status", slog.Bool("is_active", isActive))

	if err := h.Query.UpdateUserActive(ctx, db.UpdateUserActiveParams{
		AccessToken: user.AccessToken,
		IsActive:    &isActive,
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user active status: %w", err))
	}

	return connect.NewResponse(&pkplanerv1.SetUserActiveResponse{}), nil
}

// GetUserInfo returns information about the authenticated user.
func (h *PKPlanerHandler) GetUserInfo(
	ctx context.Context,
	req *connect.Request[pkplanerv1.GetUserInfoRequest],
) (*connect.Response[pkplanerv1.GetUserInfoResponse], error) {
	user := ctx.Value(AuthUserKey).(db.User)

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
