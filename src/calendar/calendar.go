package calendar

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/extractor"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func createService(ctx context.Context, token *oauth2.Token) (*calendar.Service, error) {
	config, err := createConfig()
	if err != nil {
		slog.Error("Failed to create OAuth config", slog.Any("error", err))
		return nil, err
	}

	client := config.Client(context.Background(), token)

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

func ImportTimetable(groupColumn int, token *oauth2.Token) {
	ctx := context.Background()

	slog.Info("Starting timetable import", slog.Int("group_column", groupColumn))

	timetable, err := extractor.GetTimetable(groupColumn)
	if err != nil {
		slog.Error("Failed to get timetable",
			slog.Int("group_column", groupColumn),
			slog.Any("error", err))
		return
	}
	srv, err := createService(ctx, token)
	if err != nil {
		slog.Error("Failed to create calendar service", slog.Any("error", err))
		return
	}
	currentEvents, err := getCurrentEvents(srv)
	if err != nil {
		slog.Error("Failed to get current events", slog.Any("error", err))
		return
	}

	// Create calendar events from the timetable
	newEvents := make(map[string]*calendar.Event)
	var createdCount, updatedCount, skippedCount int
	for day, entry := range timetable {
		slog.Debug("Processing day", slog.String("day", day.String()))
		for _, class := range entry {
			event := class.ToEvent()
			newEvents[event.ExtendedProperties.Private["event_id"]] = &event

			found, different, existingEvent := checkForExistingEvent(event, currentEvents)
			if found {
				if different {
					slog.Info("Updating event",
						slog.String("summary", event.Summary))
					_, err := srv.Events.Update("primary", existingEvent.Id, &event).Do()
					if err != nil {
						slog.Error("Failed to update event",
							slog.String("summary", event.Summary),
							slog.Any("error", err))
					} else {
						updatedCount++
					}
				}
				skippedCount++
				continue
			}
			// Create the event
			slog.Info("Creating event", slog.String("summary", event.Summary))
			_, err := srv.Events.Insert("primary", &event).Do()
			if err != nil {
				slog.Error("Failed to create event",
					slog.String("summary", event.Summary),
					slog.Any("error", err))
			} else {
				createdCount++
			}
		}
	}
	eventsToDelete := findEventsToDelete(currentEvents, newEvents)
	var deletedCount int
	for _, event := range eventsToDelete {
		slog.Info("Deleting event", slog.String("summary", event.Summary))
		err := srv.Events.Delete("primary", event.Id).Do()
		if err != nil {
			slog.Error("Failed to delete event",
				slog.String("summary", event.Summary),
				slog.Any("error", err))
		} else {
			deletedCount++
		}
	}

	slog.Info("Timetable import completed",
		slog.Int("created", createdCount),
		slog.Int("updated", updatedCount),
		slog.Int("deleted", deletedCount),
		slog.Int("skipped", skippedCount))
}

func createConfig() (*oauth2.Config, error) {
	b, err := os.ReadFile("credentials.json")
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

func GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
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
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read user info response", slog.Any("error", err))
		return nil, err
	}
	var result UserInfo
	if err := json.Unmarshal(data, &result); err != nil {
		slog.Error("Failed to unmarshal user info", slog.Any("error", err))
		return nil, err
	}

	slog.Info("Retrieved user info",
		slog.String("email", result.Email),
		slog.String("name", result.Name))
	return &result, nil
}

func getCurrentEvents(srv *calendar.Service) ([]*calendar.Event, error) {
	//extendedProperties has a private field "added_by" with value "GoogleCalendarTimetableImporter by Gami & Frytak"
	events, err := srv.Events.List("primary").ShowDeleted(false).OrderBy("startTime").SingleEvents(true).Do()
	if err != nil {
		slog.Error("Failed to retrieve calendar events", slog.Any("error", err))
		return nil, err
	}

	// Filter events by extendedProperties
	filteredEvents := make([]*calendar.Event, 0)
	for _, event := range events.Items {
		if event.ExtendedProperties != nil && event.ExtendedProperties.Private != nil {
			if event.ExtendedProperties.Private["added_by"] == extractor.ADDED_BY {
				filteredEvents = append(filteredEvents, event)
			}
		}
	}
	slog.Debug("Retrieved current events", slog.Int("count", len(filteredEvents)))
	return filteredEvents, nil
}

// can return that it found one and its identical, or that it found one but it's different, or that it found none, and can return the found event if any
// returns: (found bool, different bool, event *calendar.Event)
func checkForExistingEvent(newEvent calendar.Event, currentEvents []*calendar.Event) (bool, bool, *calendar.Event) {
	for _, event := range currentEvents {
		if event.ExtendedProperties.Private["event_id"] == newEvent.ExtendedProperties.Private["event_id"] {
			if event.ExtendedProperties.Private["hash"] != newEvent.ExtendedProperties.Private["hash"] {
				slog.Debug("Event differs",
					slog.String("old_summary", event.Summary),
					slog.String("new_summary", newEvent.Summary))
				return true, true, event
			}
			slog.Debug("Event is identical", slog.String("summary", event.Summary))
			return true, false, event
		}
	}
	slog.Debug("Event not found", slog.String("summary", newEvent.Summary))
	return false, false, nil
}

func findEventsToDelete(currentEvents []*calendar.Event, newEvents map[string]*calendar.Event) []*calendar.Event {
	eventsToDelete := make([]*calendar.Event, 0)
	for _, event := range currentEvents {
		if _, ok := newEvents[event.ExtendedProperties.Private["event_id"]]; !ok {
			eventsToDelete = append(eventsToDelete, event)
		}
	}
	return eventsToDelete
}
