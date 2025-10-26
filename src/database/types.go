package database

import "time"

type User struct {
	ID           *int
	Email        *string
	ChosenColumn *int
	Name         *string
	GivenName    *string
	Picture      *string

	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
	ExpiresIn    int64
}

type AuthorizeResponse struct {
	AccessToken  string `json:"access_token"`
	Email        string `json:"email"`
	ChosenColumn int    `json:"chosen_column"`
	Name         string `json:"name"`
	Picture      string `json:"picture"`
}
