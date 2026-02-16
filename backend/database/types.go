package database

import (
	"time"

	"golang.org/x/oauth2"
)

type User struct {
	ID           *int64
	Email        *string
	ChosenColumn *int64
	IsActive     *bool
	Name         *string
	GivenName    *string
	Picture      *string
	LastSynced   *time.Time

	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
	ExpiresIn    int64
}

func (u *User) Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  u.AccessToken,
		TokenType:    u.TokenType,
		RefreshToken: u.RefreshToken,
		Expiry:       u.Expiry,
		ExpiresIn:    u.ExpiresIn,
	}
}

type AuthorizeResponse struct {
	AccessToken  string `json:"access_token"`
	Email        string `json:"email"`
	ChosenColumn int64  `json:"chosen_column"`
	Name         string `json:"name"`
	Picture      string `json:"picture"`
}

type UserInfoResponse struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	GivenName    string `json:"given_name"`
	Picture      string `json:"picture"`
	IsActive     bool   `json:"is_active"`
	ChosenColumn int64  `json:"chosen_column"`
}
