package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/WoodPKrs/GoogleCalendarTimetable/database/gen"
	"golang.org/x/oauth2"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

// Queries wraps the sqlc-generated gen.Queries and maps results to domain types.
type Queries struct {
	gen *gen.Queries
}

func New(db *sql.DB) *Queries {
	return &Queries{gen: gen.New(db)}
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{gen: q.gen.WithTx(tx)}
}

// Probably a big no no, but i didnt want to figure out how to pass it to every function that needs it, but it also should work just fine like this
var Query *Queries

// InitializeDatabase creates tables (if they don't exist) and sets up the global Query.
func InitializeDatabase(db *sql.DB) error {
	if _, err := db.Exec(ddl); err != nil {
		return err
	}
	Query = New(db)
	return nil
}

// ---------------------------------------------------------------------------
// Query wrappers — same signatures as before, delegates to gen.Queries
// ---------------------------------------------------------------------------

func (q *Queries) UpsertUser(user *User) (User, error) {
	ctx := context.Background()
	row, err := q.gen.UpsertUser(ctx, gen.UpsertUserParams{
		Email:        user.Email,
		Name:         user.Name,
		GivenName:    user.GivenName,
		Picture:      user.Picture,
		AccessToken:  user.AccessToken,
		TokenType:    user.TokenType,
		RefreshToken: user.RefreshToken,
		Expiry:       user.Expiry,
		ExpiresIn:    user.ExpiresIn,
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return User{}, err
	}
	return User{
		Email:        row.Email,
		Name:         row.Name,
		GivenName:    row.GivenName,
		Picture:      row.Picture,
		AccessToken:  row.AccessToken,
		TokenType:    row.TokenType,
		RefreshToken: row.RefreshToken,
		Expiry:       row.Expiry,
		ExpiresIn:    row.ExpiresIn,
		ChosenColumn: row.ChosenColumn,
	}, nil
}

func (q *Queries) GetUserByEmail(email string) (*User, error) {
	ctx := context.Background()
	row, err := q.gen.GetUserByEmail(ctx, &email)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}

	return &User{
		ID:           &row.ID,
		Email:        row.Email,
		Name:         row.Name,
		GivenName:    row.GivenName,
		Picture:      row.Picture,
		AccessToken:  row.AccessToken,
		TokenType:    row.TokenType,
		RefreshToken: row.RefreshToken,
		Expiry:       row.Expiry,
		ExpiresIn:    row.ExpiresIn,
		ChosenColumn: row.ChosenColumn,
	}, nil
}

func (q *Queries) GetAllUsers() ([]*User, error) {
	ctx := context.Background()
	rows, err := q.gen.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*User, 0, len(rows))
	for _, row := range rows {
		lastSynced := row.LastSynced
		users = append(users, &User{
			ID:           &row.ID,
			Email:        row.Email,
			Name:         row.Name,
			GivenName:    row.GivenName,
			Picture:      row.Picture,
			AccessToken:  row.AccessToken,
			TokenType:    row.TokenType,
			RefreshToken: row.RefreshToken,
			Expiry:       row.Expiry,
			ExpiresIn:    row.ExpiresIn,
			ChosenColumn: row.ChosenColumn,
			IsActive:     row.IsActive,
			LastSynced:   &lastSynced,
		})
	}
	return users, nil
}

func (q *Queries) UpdateChosenColumn(accessToken string, chosenColumn int64) error {
	ctx := context.Background()
	err := q.gen.UpdateChosenColumn(ctx, gen.UpdateChosenColumnParams{
		ChosenColumn: &chosenColumn,
		AccessToken:  accessToken,
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	return nil
}

func (q *Queries) GetTokenByAccessToken(accessToken string) (*oauth2.Token, error) {
	ctx := context.Background()
	row, err := q.gen.GetTokenByAccessToken(ctx, accessToken)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	return &oauth2.Token{
		AccessToken:  row.AccessToken,
		TokenType:    row.TokenType,
		RefreshToken: row.RefreshToken,
		Expiry:       row.Expiry,
		ExpiresIn:    row.ExpiresIn,
	}, nil
}

func (q *Queries) GetUserByAccessToken(accessToken string) (*User, error) {
	ctx := context.Background()
	row, err := q.gen.GetUserByAccessToken(ctx, accessToken)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	return &User{
		Email:        row.Email,
		Name:         row.Name,
		GivenName:    row.GivenName,
		Picture:      row.Picture,
		AccessToken:  row.AccessToken,
		TokenType:    row.TokenType,
		RefreshToken: row.RefreshToken,
		Expiry:       row.Expiry,
		ExpiresIn:    row.ExpiresIn,
		ChosenColumn: row.ChosenColumn,
		IsActive:     row.IsActive,
	}, nil
}

func (q *Queries) UpdateUserSyncTime(accessToken string) error {
	ctx := context.Background()
	err := q.gen.UpdateUserSyncTime(ctx, accessToken)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	return nil
}
