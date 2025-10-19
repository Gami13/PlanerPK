package database

import (
	"context"
	"database/sql"
	"fmt"

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

type EndpointContext struct {
	Query *Queries
	Ctx   context.Context
}

func NewEndpointContext(query *Queries, ctx context.Context) *EndpointContext {
	if query == nil {
		panic("Query is nil!")
	}
	if ctx == nil {
		panic("Context is nil!")
	}

	return &EndpointContext{Query: query, Ctx: ctx}
}

// just creates the chud tables if they dont exist yet
func InitializeDatabase(db *sql.DB) error {
	query := New(db)
	err := query.createTables()
	if err != nil {
		return err
	}
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
		chosen_column INTEGER,
		is_active BOOLEAN DEFAULT 1
	);
	`
	_, err := q.db.Exec(createUserTable)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) UpsertChudUser(user *User) error {
	upsertUser := `
	INSERT INTO users (email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(email) DO UPDATE SET
		name = excluded.name,
		given_name = excluded.given_name,
		picture = excluded.picture,
		access_token = excluded.access_token,
		token_type = excluded.token_type,
		refresh_token = excluded.refresh_token,
		expiry = excluded.expiry,
		expires_in = excluded.expires_in,
		chosen_column = excluded.chosen_column;
	`
	_, err := q.db.Exec(upsertUser, user.Email, user.Name, user.GivenName, user.Picture, user.AccessToken, user.TokenType, user.RefreshToken, user.Expiry, user.ExpiresIn, user.ChosenColumn)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	return nil
}

func (q *Queries) GetChudUserByEmail(email string) (*User, error) {
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

func (q *Queries) GetAllChudUsers() ([]*User, error) {
	getUsers := `
	SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column
	FROM users;
	`
	rows, err := q.db.Query(getUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.GivenName, &user.Picture, &user.AccessToken, &user.TokenType, &user.RefreshToken, &user.Expiry, &user.ExpiresIn, &user.ChosenColumn); err != nil {
			return nil, err
		}
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
