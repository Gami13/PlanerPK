package handler

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/calendar"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pkplanerv1 "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
)

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
