package database

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	_ "modernc.org/sqlite"
)

type DBTX interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

// Probably a big no no, but i didnt want to figure out how to pass it to every function that needs it, but it also should work just fine like this
var Query *Queries

// just creates the chud tables if they dont exist yet
func InitializeDatabase(db *sql.DB) error {
	query := New(db)
	err := query.createTables()
	if err != nil {
		return err
	}
	Query = query
	return nil
}

func (q *Queries) createTables() error {
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		given_name TEXT,
		picture TEXT,
		access_token TEXT NOT NULL,
		token_type TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expiry TIMESTAMP NOT NULL,
		expires_in INTEGER NOT NULL,
		chosen_column INTEGER DEFAULT -1,
		is_active BOOLEAN DEFAULT 1,
		last_synced TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);
	`
	_, err := q.db.Exec(createUserTable)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) UpsertUser(user *User) (User, error) {
	//should probably return all user fields
	upsertUser := `
	INSERT INTO users (email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(email) DO UPDATE SET
		name = excluded.name,
		given_name = excluded.given_name,
		picture = excluded.picture,
		access_token = excluded.access_token,
		token_type = excluded.token_type,
		refresh_token = excluded.refresh_token,
		expiry = excluded.expiry,
		expires_in = excluded.expires_in
	RETURNING email,name,given_name,picture,access_token,token_type,refresh_token,expiry,expires_in,chosen_column;
	`
	var returnedUser User
	err := q.db.QueryRow(upsertUser, user.Email, user.Name, user.GivenName, user.Picture, user.AccessToken, user.TokenType, user.RefreshToken, user.Expiry, user.ExpiresIn).Scan(&returnedUser.Email, &returnedUser.Name, &returnedUser.GivenName, &returnedUser.Picture, &returnedUser.AccessToken, &returnedUser.TokenType, &returnedUser.RefreshToken, &returnedUser.Expiry, &returnedUser.ExpiresIn, &returnedUser.ChosenColumn)
	if err != nil {
		fmt.Println("Error: ", err)
		return User{}, err
	}
	return returnedUser, nil
}

func (q *Queries) GetUserByEmail(email string) (*User, error) {
	getUser := `
	SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column
	FROM users
	WHERE email = ?;
	`
	row := q.db.QueryRow(getUser, email)
	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.GivenName, &user.Picture, &user.AccessToken, &user.TokenType, &user.RefreshToken, &user.Expiry, &user.ExpiresIn, &user.ChosenColumn)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	return &user, nil
}

func (q *Queries) GetAllUsers() ([]*User, error) {
	getUsers := `
	SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, is_active, last_synced
	FROM users;
	`
	rows, err := q.db.Query(getUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var date string
	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.GivenName, &user.Picture, &user.AccessToken, &user.TokenType, &user.RefreshToken, &user.Expiry, &user.ExpiresIn, &user.ChosenColumn, &user.IsActive, &date); err != nil {
			return nil, err
		}
		newTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return nil, err
		}
		user.LastSynced = &newTime
		users = append(users, &user)
	}
	return users, nil
}

func (q *Queries) UpdateChosenColumn(accessToken string, chosenColumn int) error {
	updateUser := `
	UPDATE users
	SET chosen_column = ?
	WHERE access_token = ?;
	`
	_, err := q.db.Exec(updateUser, chosenColumn, accessToken)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	return nil
}

func (q *Queries) GetTokenByAccessToken(accessToken string) (*oauth2.Token, error) {
	getUser := `
	SELECT access_token, token_type, refresh_token, expiry, expires_in
	FROM users
	WHERE access_token = ?;
	`
	row := q.db.QueryRow(getUser, accessToken)
	var token oauth2.Token
	err := row.Scan(&token.AccessToken, &token.TokenType, &token.RefreshToken, &token.Expiry, &token.ExpiresIn)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	return &token, nil
}

func (q *Queries) GetUserByAccessToken(accessToken string) (*User, error) {
	getUser := `
	SELECT email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, is_active
	FROM users
	WHERE access_token = ?;
	`
	row := q.db.QueryRow(getUser, accessToken)
	var user User
	err := row.Scan(&user.Email, &user.Name, &user.GivenName, &user.Picture, &user.AccessToken, &user.TokenType, &user.RefreshToken, &user.Expiry, &user.ExpiresIn, &user.ChosenColumn, &user.IsActive)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	return &user, nil
}

func (q *Queries) UpdateUserSyncTime(accessToken string) error {
	updateSyncTime := `
	UPDATE users
	SET last_synced = (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	WHERE access_token = ?;
	`
	_, err := q.db.Exec(updateSyncTime, accessToken)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	return nil
}
