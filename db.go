package main

import (
	"database/sql"
	"fmt"
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

	// Users table
	db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	full_name TEXT UNIQUE,
    	hashed_password TEXT NOT NULL,
    	class TEXT
		points INTEGER DEFAULT 0
	)
	`)

	// Sessions table
	db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user_id INTEGER NOT NULL,
    	session_token TEXT UNIQUE NOT NULL,
    	expires_at DATETIME,
    	FOREIGN KEY(user_id) REFERENCES users(id)
	)
	`)

	// Questiosn table
	db.Exec(`
	CREATE TABLE IF NOT EXISTS questions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		question TEXT NOT NULL,
		subject TEXT,
		for_grade TEXT,
		option_a TEXT,
		option_b TEXT,
		option_c TEXT,
		option_d TEXT,
		correct_answer TEXT,
		points_worth INT
	)
	`)

	// User question history
	db.Exec(`
	CREATE TABLE IF NOT EXISTS user_question_history (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user_id INTEGER,
    	question_id INTEGER,
    	is_correct INTEGER,
    	points_earned INTEGER,
    	answered_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    	UNIQUE(user_id, question_id)
	);
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
	).Scan(&usr.Id, &usr.FullName, &usr.hashedPassword, &usr.Grade)

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
	_, err = db.Exec("INSERT INTO users (full_name, hashed_password, class) VALUES (?, ?, ?)", usr.FullName, usr.hashedPassword, usr.Grade)
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
	err = db.QueryRow("SELECT id, full_name, hashed_password, class, points FROM users WHERE id = ?", usr_id).
		Scan(&usr.Id, &usr.FullName, &usr.hashedPassword, &usr.Grade, &usr.Points)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &usr, nil
}

func QueryQuestionsForUser(user_id int, subject string, grade string) ([]Question, error) {
	rows, err := db.Query(`SELECT id, question, option_a, option_b, option_c, option_d, correct_answer, points_worth FROM questions 
		WHERE NOT EXISTS (
			SELECT 1 FROM user_question_history h 
			WHERE h.user_id = ? AND h.question_id = questions.id
		)
		AND subject = ? 
		AND for_grade = ?
		ORDER BY RANDOM()
		LIMIT 10`, user_id, subject, grade)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question

		err := rows.Scan(&q.Id, &q.Question, &q.A, &q.B, &q.C, &q.D, &q.Correct_answer, &q.Points_worth)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func AddQuestionToDB(q Question) error {
	_, err := db.Exec(
		`INSERT INTO questions (question, subject, for_grade, option_a, option_b, option_c, option_d, correct_answer, points_worth)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		q.Question, q.Subject, q.For_grade, q.A, q.B, q.C, q.D, q.Correct_answer, q.Points_worth,
	)
	if err != nil {
		return fmt.Errorf("failed adding question to database: %w", err)
	}
	return nil
}

func CheckIfAnswerCorrect(question_id int, answer string) (bool, int, error) {
	var real_answer string
	var points_worth int
	err := db.QueryRow("SELECT correct_answer, points_worth FROM questions WHERE id = ?", question_id).Scan(&real_answer, &points_worth)
	if err != nil {
		return false, 0, fmt.Errorf("failed checking answer: %w", err)
	}

	return real_answer == answer, points_worth, nil
}

func UpdateUserPoints(usr *User, total_points int) error {
	_, err := db.Exec("UPDATE users SET points = points + ? WHERE id = ?", total_points, usr.Id)
	usr.Points += total_points
	if err != nil {
		return err
	}
	return nil
}

func GetQuestionFromID(question_id int) (Question, error) {
	var q Question
	err := db.QueryRow("SELECT question, correct_answer, points_worth FROM questions WHERE id = ?", question_id).Scan(&q.Question, &q.Correct_answer, &q.Points_worth)
	if err != nil {
		return q, err
	}
	return q, nil
}

func UserDoneQuestion(user_id int, q Question, answer string, qId int) error {
	is_correct := false
	points := 0
	if answer == q.Correct_answer {
		is_correct = true
		points = q.Points_worth
	}

	_, err := db.Exec(`INSERT INTO user_question_history (user_id, question_id, is_correct, points_earned) VALUES (?, ?, ?, ?) ON CONFLICT(user_id, question_id) DO NOTHING`, user_id, qId, is_correct, points)
	if err != nil {
		return err
	}
	return nil
}
