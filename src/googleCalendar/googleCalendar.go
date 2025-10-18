//thist is copied from the docs btw

package googleCalendar

import (
	"GoogleCalendarTimetable/src/timetableExtractor"
	"GoogleCalendarTimetable/src/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Retrieves a token from a local file.
func tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open("token.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func createService(ctx context.Context) (*calendar.Service, error) {
	config, err := createConfig()
	if err != nil {
		log.Fatalf("Unable to create config: %v", err)
		return nil, err
	}
	credentials, err := tokenFromFile()
	if err != nil {
		log.Fatalf("Unable to get token from file: %v", err)
		return nil, err
	}
	client := config.Client(context.Background(), credentials)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
		return nil, err
	}
	return srv, nil
}

func ImportTimetable(config *utils.Config, groupColumn int) {
	ctx := context.Background()

	srv, err := createService(ctx)
	if err != nil {
		log.Fatalf("Unable to create calendar service: %v", err)
	}

	timetable := timetableExtractor.GetTimetable(config, groupColumn)

	// Create calendar events from the timetable
	for day, entry := range timetable {
		// Name (type) - Teacher | Location | AdditionalInfo
		fmt.Println("Inserting day: ", day)
		dayDate := utils.GetDateOfNextWeekday(day)
		for _, class := range entry {
			startTime := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), class.StartTime.Hour(), class.StartTime.Minute(), 0, 0, time.Now().Local().Location())
			endTime := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), class.EndTime.Hour(), class.EndTime.Minute(), 0, 0, time.Now().Local().Location())

			summary := fmt.Sprintf("%s (%s) - %s | %s", class.Name, class.Type, class.Teacher, class.Location)
			if class.AdditionalInfo != "" {
				summary += fmt.Sprintf(" | %s", class.AdditionalInfo)
			}

			event := &calendar.Event{
				Summary: summary,
				Start: &calendar.EventDateTime{
					DateTime: startTime.Format(time.RFC3339),
				},
				End: &calendar.EventDateTime{
					DateTime: endTime.Format(time.RFC3339),
				},
			}
			_, err := srv.Events.Insert("primary", event).Do()
			if err != nil {
				log.Printf("Unable to create event: %v", err)
			}
		}

	}

}

func createConfig() (*oauth2.Config, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
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
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%s\n", authURL)

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	saveToken("token.json", tok)
	return tok, nil
}
