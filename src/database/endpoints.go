package database

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/calendar"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/types"
)

// PUT /setChosenColumn
// in body: chosenColumn: int
// sets the chosen column for the user associated with the access token in the Authorization header
func (ec *EndpointContext) SetChosenColumnEndpoint(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	var requestBody struct {
		ChosenColumn int `json:"chosenColumn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return
	}
	if err := ec.Query.UpdateChosenColumn(accessToken, requestBody.ChosenColumn); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update chosen column: %v", err), http.StatusInternalServerError)
		return
	}

	//Update the timetable
	//TODO: This is currently a temporary solution, as it makes the endpoint slow, In the future, we should have a non-blocking way to update the timetable
	config := types.NewConfig()
	timetablePath, err := extractor.GetCurrentTimetable()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get current timetable: %v", err), http.StatusInternalServerError)
		return
	}
	timetable := extractor.GetTimetable(config, timetablePath+"x", requestBody.ChosenColumn)
	token, err := ec.Query.GetTokenByAccessToken(accessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get token: %v", err), http.StatusInternalServerError)
		return
	}
	calendar.ImportTimetable(config, timetable, token)
	w.WriteHeader(http.StatusOK)
}

func (ec *EndpointContext) AuthorizeEndpoint(w http.ResponseWriter, r *http.Request) {

	//Get the oauth code from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	tok, err := calendar.AuthorizeFromCode(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Authorization failed: %v", err), http.StatusInternalServerError)
		return
	}

	userInfo, err := calendar.GetUserInfo(tok)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}
	user := User{
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
	err = ec.Query.UpsertUser(&user)
	if err != nil {
		fmt.Printf("Failed to save user: %v\n", err)
		return
	}

	fmt.Println("UwU")
	w.Write([]byte(user.AccessToken))
}
