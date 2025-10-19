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
