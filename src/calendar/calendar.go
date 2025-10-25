package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func createService(ctx context.Context, token *oauth2.Token) (*calendar.Service, error) {
	config, err := createConfig()
	if err != nil {
		log.Fatalf("Unable to create config: %v", err)
		return nil, err
	}

	client := config.Client(context.Background(), token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
		return nil, err
	}
	return srv, nil
}

func ImportTimetable(config *types.Config, timetable extractor.Timetable, token *oauth2.Token) {
	ctx := context.Background()

	srv, err := createService(ctx, token)
	if err != nil {
		log.Fatalf("Unable to create calendar service: %v", err)
	}
	currentEvents, err := getCurrentEvents(srv)
	if err != nil {
		log.Fatalf("Unable to get current events: %v", err)
	}

	// Create calendar events from the timetable
	newEvents := make(map[string]*calendar.Event)
	for day, entry := range timetable {
		// Name (type) - Teacher | Location | AdditionalInfo
		fmt.Println("Inserting day: ", day)
		for _, class := range entry {
			event := class.ToEvent()
			newEvents[event.ExtendedProperties.Private["event_id"]] = &event

			found, different, existingEvent := checkForExistingEvent(event, currentEvents)
			if found {
				if different {
					// Update the event
					fmt.Println("Updating event")
					_, err := srv.Events.Update("primary", existingEvent.Id, &event).Do()
					if err != nil {
						log.Printf("Unable to update event: %v", err)
					}
				}
				fmt.Println("Event already exists, skipping creation")
				continue
			}
			// Create the event

			fmt.Println("Creating event: ", event.Summary)
			_, err := srv.Events.Insert("primary", &event).Do()
			if err != nil {
				log.Printf("Unable to create event: %v", err)
			}
		}

	}
	eventsToDelete := findEventsToDelete(currentEvents, newEvents)
	for _, event := range eventsToDelete {
		fmt.Println("Deleting event: ", event.Summary)
		err := srv.Events.Delete("primary", event.Id).Do()
		if err != nil {
			log.Printf("Unable to delete event: %v", err)
		}
	}

}

func createConfig() (*oauth2.Config, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	scopes := []string{calendar.CalendarScope, "email", "profile"}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config, nil
}

func AuthorizeFromCode(code string) (*oauth2.Token, error) {
	config, err := createConfig()
	if err != nil {
		return nil, err
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%s\n", authURL+"&prompt=consent")

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	config, err := createConfig()
	if err != nil {
		log.Fatalf("Unable to create config: %v", err)
	}
	client := config.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		fmt.Println("Error getting user info: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading user info response: ", err)
		return nil, err
	}
	var result UserInfo
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Println("Error unmarshalling user info response: ", err)
		return nil, err
	}

	fmt.Println("Got Chud data: ", result.String())
	return &result, nil

}

func getCurrentEvents(srv *calendar.Service) ([]*calendar.Event, error) {
	//extendedProperties has a private field "added_by" with value "GoogleCalendarTimetableImporter by Gami & Frytak"
	events, err := srv.Events.List("primary").ShowDeleted(false).OrderBy("startTime").SingleEvents(true).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
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
	return filteredEvents, nil
}

// can return that it found one and its identical, or that it found one but it's different, or that it found none, and can return the found event if any
// returns: (found bool, different bool, event *calendar.Event)
// ! Currently does not detect if an event was deleted from the source (sheet) and should be deleted from the calendar as well
func checkForExistingEvent(newEvent calendar.Event, currentEvents []*calendar.Event) (bool, bool, *calendar.Event) {
	for _, event := range currentEvents {

		if event.ExtendedProperties.Private["event_id"] == newEvent.ExtendedProperties.Private["event_id"] {
			if event.ExtendedProperties.Private["hash"] != newEvent.ExtendedProperties.Private["hash"] {
				fmt.Println("Event differs: ", event.Summary, " vs ", newEvent.Summary)
				return true, true, event
			}
			fmt.Println("Event is identical: ", event.Summary)
			return true, false, event
		}
	}
	fmt.Println("Event not found: ", newEvent.Summary)
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
