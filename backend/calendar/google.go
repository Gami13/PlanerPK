package calendar

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"

	"github.com/woodpkrs/PlanerPK/global"
	"github.com/woodpkrs/PlanerPK/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func createConfig() (*oauth2.Config, error) {
	b, err := os.ReadFile(global.CalendarCredentialsPath())
	if err != nil {
		slog.Error("Failed to read credentials file", slog.Any("error", err))
		return nil, err
	}

	scopes := []string{calendar.CalendarScope, "email", "profile"}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		slog.Error("Failed to parse credentials file", slog.Any("error", err))
		return nil, err
	}
	return config, nil
}

func AuthorizeFromCode(code string) (*oauth2.Token, error) {
	config, err := createConfig()
	if err != nil {
		return nil, err
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func GetUserInfo(token *oauth2.Token) (*types.UserInfo, error) {
	config, err := createConfig()
	if err != nil {
		slog.Error("Failed to create config for user info", slog.Any("error", err))
		return nil, err
	}
	client := config.Client(context.Background(), token)

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
	config, err := createConfig()
	if err != nil {
		slog.Error("Failed to create OAuth config", slog.Any("error", err))
		return nil, err
	}

	client := config.Client(ctx, token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		slog.Error("Failed to create calendar service", slog.Any("error", err))
		return nil, err
	}
	return srv, nil
}

func GenerateAuthURL() (string, error) {
	config, err := createConfig()
	if err != nil {
		return "", err
	}
	//"&prompt=consent"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	return authURL, nil
}
