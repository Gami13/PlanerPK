package calendar

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/woodpkrs/PlanerPK/global"
	"github.com/woodpkrs/PlanerPK/types"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func AuthorizeFromCode(code string) (*oauth2.Token, error) {
	tok, err := global.Config.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func GetUserInfo(token *oauth2.Token) (*types.UserInfo, error) {
	client := global.Config.OAuthConfig.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		slog.Error("Failed to get user info from Google", slog.Any("error", err))
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read user info response", slog.Any("error", err))
		return nil, err
	}
	var result types.UserInfo
	if err := json.Unmarshal(data, &result); err != nil {
		slog.Error("Failed to unmarshal user info", slog.Any("error", err))
		return nil, err
	}

	slog.Info("Retrieved user info",
		slog.String("email", result.Email),
		slog.String("name", result.Name))
	return &result, nil
}

func createService(ctx context.Context, token *oauth2.Token) (*calendar.Service, error) {
	client := global.Config.OAuthConfig.Client(ctx, token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		slog.Error("Failed to create calendar service", slog.Any("error", err))
		return nil, err
	}
	return srv, nil
}

func GenerateAuthURL() (string, error) {
	//"&prompt=consent"
	authURL := global.Config.OAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	return authURL, nil
}
