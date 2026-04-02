package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db  *sql.DB
	err error
)

func InitDB() {
	db, err = sql.Open("sqlite3", "users.db")
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	full_name TEXT UNIQUE,
    	hashed_password TEXT NOT NULL,
    	class TEXT
	)
	`)
	db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user_id INTEGER NOT NULL,
    	session_token TEXT UNIQUE NOT NULL,
    	expires_at DATETIME,
    	FOREIGN KEY(user_id) REFERENCES users(id)
	)
	`)
}

func GetDB() *sql.DB {
	return db
}

func CheckUserInDB(user string) (bool, error) {
	var id int

	err := db.QueryRow(
		"SELECT id FROM users WHERE full_name = ?",
		user,
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // user does NOT exist
		}
		return false, err // real error
	}

	return true, nil // user exists
}

func GetUser(user string) (*User, error) {
	var usr User

	err := db.QueryRow(
		"SELECT id, full_name, hashed_password, class FROM users WHERE full_name = ?",
		user,
	).Scan(&usr.id, &usr.fullName, &usr.hashedPassword, &usr.grade)

	if err != nil {
		return nil, err
	}

	return &usr, nil
}

func AddSessionToDB(userID int, sessionToken string) error {
	_, err := db.Exec(
		"INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, datetime('now', '+24 hours'))",
		userID, sessionToken,
	)
	return err
}

func AddUserToDB(usr *User) error {
	_, err = db.Exec("INSERT INTO users (full_name, hashed_password, class) VALUES (?, ?, ?)", usr.fullName, usr.hashedPassword, usr.grade)
	return err
}

func GetUserFromSession(session string) (*User, error) {
	var usr_id int
	var usr User

	// Get user ID from session
	err := db.QueryRow("SELECT user_id FROM sessions WHERE session_token = ? AND expires_at > datetime('now')", session).Scan(&usr_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get user info from users table
	err = db.QueryRow("SELECT id, full_name, hashed_password, class FROM users WHERE id = ?", usr_id).
		Scan(&usr.id, &usr.fullName, &usr.hashedPassword, &usr.grade)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &usr, nil
}
