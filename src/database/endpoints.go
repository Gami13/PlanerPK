package database

import (
	"GoogleCalendarTimetable/src/googleCalendar"
	"GoogleCalendarTimetable/src/utils"
	"fmt"
	"net/http"
)

func (ec *EndpointContext) AuthorizeEndpoint(w http.ResponseWriter, r *http.Request) {

	//Get the oauth code from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	fmt.Println("Code: ", code)
	googleCalendar.AuthorizeFromCode(code)
	config := utils.NewConfig()

	googleCalendar.ImportTimetable(config, 3)

	fmt.Println("UwU")
	w.Write([]byte("UwU"))
}
