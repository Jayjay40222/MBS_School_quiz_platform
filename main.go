package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	InitDB()
	que := Question{Question: "What is Atomic mass?", Subject: "Chemistry", For_grade: "Grade 9", A: "Some some", B: "b soe som", C: "c som", D: "d som", Correct_answer: "A", Points_worth: 5}
	AddQuestionToDB(que)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/login/", handleLogin)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/signup/", handleSignup)
	mux.HandleFunc("/signup", handleSignup)
	mux.HandleFunc("/dashboard", handleDashboard)
	mux.HandleFunc("/quiz", handleQuiz)
	mux.HandleFunc("/quiz/", handleQuiz)
	mux.HandleFunc("/quiz/play", handleQuizPlay)
	mux.HandleFunc("/quiz/submit", handleQuizSubmit)
	mux.HandleFunc("/teacher_login/", handleTeacherLogin)
	mux.HandleFunc("/teacher_login", handleTeacherLogin)
	log.Fatal(http.ListenAndServe(":8080", mux))
	log.Println("Server started on http://localhost:8080")
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_token")
	if err == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
func handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "templates/login.html")
		return
	case http.MethodPost:
		// Store creds
		username := r.FormValue("name")
		password := r.FormValue("pass")
		class := r.FormValue("cls")

		// Check if user exists in DB
		exists, err := CheckUserInDB(username)
		if err != nil {
			// Server error
			log.Fatal("Failed checking user in DB")
			http.Error(w, "Server Error", 500)
		}

		if exists {
			usr, err := GetUser(username)
			if err != nil {
				// Server error
				log.Fatal("Failed getting user from DB")
				http.Error(w, "Server Error", 500)
			}

			if err := bcrypt.CompareHashAndPassword([]byte(usr.hashedPassword), []byte(password)); err != nil || usr.Grade != class {
				w.Write([]byte("Incorrect password or class"))
				return
			}

			// Generate session token
			sessionToken, err := generateSessionToken()
			if err != nil {
				log.Fatal("Failed generating session token")
				http.Error(w, "Server Error", 500)
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    sessionToken,
				HttpOnly: true,
				Secure:   false,
				Path:     "/",
			})

			err = AddSessionToDB(usr.Id, sessionToken)
			if err != nil {
				log.Fatal("Failed adding session to DB")
			}

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			log.Println("Session token generated:", sessionToken)
			return
		} else {
			w.Write([]byte("User does not exist, sign up instead"))
		}
	}
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "templates/signup.html")
		return
	case http.MethodPost:
		username := r.FormValue("name")
		password := r.FormValue("pass")
		class := r.FormValue("cls")

		// Check if user already exists
		exists, err := CheckUserInDB(username)
		if err != nil {
			log.Println("DB error:", err)
			http.Error(w, "Server error", 500)
			return
		}
		if exists {
			http.Error(w, "Username already taken", http.StatusBadRequest)
			return
		}

		// Confirm password strength
		if len(password) < 8 {
			http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
			return
		}

		hashPass, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			http.Error(w, "Server error", 500)
			return
		}

		usr := &User{
			FullName:       username,
			hashedPassword: string(hashPass),
			Grade:          class,
		}

		err = AddUserToDB(usr)
		if err != nil {
			log.Println("DB error:", err)
			http.Error(w, "Failed to create account", 500)
			return
		}

		log.Println("Account created")

		// Redirect after signup (no w.Write)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func getCurrentUser(w http.ResponseWriter, r *http.Request) (*User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Println("No session cookie found. Redirecting to /login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, err
	}

	usr, err := GetUserFromSession(cookie.Value)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return nil, err
	}

	if usr == nil {
		log.Println("Invalid session. Redirecting to /login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, err
	}
	return usr, nil
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	usr, err := getCurrentUser(w, r)
	if err != nil {
		return
	}

	// User is authenticated — serve dashboard
	err = templates.ExecuteTemplate(w, "dashboard.html", usr)
	if err != nil {
		log.Println("Template execution error:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}

func handleQuiz(w http.ResponseWriter, r *http.Request) {
	usr, err := getCurrentUser(w, r)
	if err != nil {
		return
	}

	switch r.Method {
	case http.MethodGet:
		err = templates.ExecuteTemplate(w, "quiz.html", usr)
		log.Println("User Points: ", usr.Points)
		if err != nil {
			log.Println("Template execution error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}

	}
}

func handleQuizPlay(w http.ResponseWriter, r *http.Request) {
	usr, err := getCurrentUser(w, r)
	if err != nil {
		return
	}

	subject := r.URL.Query().Get("subject")
	grade := r.URL.Query().Get("grade")

	// Query questions for user
	questions, err := QueryQuestionsForUser(usr.Id, subject, grade)
	if err != nil {
		http.Error(w, "Server error", 500)
		log.Fatal("Failed querying questions", err)
		return
	}

	quizData := QuizPageData{User: *usr, Questions: questions}

	fmt.Println("Sub:", subject, "Grade:", grade)

	switch r.Method {
	case http.MethodGet:
		err = templates.ExecuteTemplate(w, "play.html", quizData)
		if err != nil {
			http.Error(w, "Server Error", 500)
			log.Println("Template error: ", err)
			return
		}
	}
}

func handleTeacherLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := templates.ExecuteTemplate(w, "teacher_login.html", nil)
		if err != nil {
			log.Println("Template execution error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

}

func handleQuizSubmit(w http.ResponseWriter, r *http.Request) {
	usr, err := getCurrentUser(w, r)
	if err != nil {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "only POST requests allowed.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}
	var results []QuestionResult
	totalPoints := 0

	for key, values := range r.Form {
		if strings.HasPrefix(key, "qid_") {
			qidStr := values[0] // the question ID
			qid, err := strconv.Atoi(qidStr)
			if err != nil {
				log.Println("Invalid question ID:", qidStr)
				continue
			}
			ans := r.FormValue("answer_" + strings.TrimPrefix(key, "qid_"))

			question, err := GetQuestionFromID(qid)
			if err != nil {
				http.Error(w, "Server Error", http.StatusInternalServerError)
				log.Println("Error fetching question:", err)
				return
			}

			err = UserDoneQuestion(usr.Id, question, ans, qid)
			if err != nil {
				http.Error(w, "Server Error", http.StatusInternalServerError)
				log.Println("Failed adding question to history:", err)
				return
			}

			fmt.Println("Question ID:", qidStr, "Answer:", ans)
			is_correct := ans == question.Correct_answer
			earned := 0
			if is_correct {
				earned = question.Points_worth
				totalPoints += earned
			}

			results = append(results, QuestionResult{
				Question:      question.Question,
				GivenAnswer:   ans,
				CorrectAnswer: question.Correct_answer,
				PointsEarned:  earned,
				IsCorrect:     is_correct,
			})
		}
	}
	// Update points
	err = UpdateUserPoints(usr, totalPoints)
	if err != nil {
		log.Println("Failed updating user points. Error:", err)
		http.Error(w, "Failed to update score", http.StatusInternalServerError)
		return
	}

	data := struct {
		User        User
		Results     []QuestionResult
		TotalPoints int
	}{
		User:        *usr,
		Results:     results,
		TotalPoints: totalPoints,
	}

	err = templates.ExecuteTemplate(w, "results.html", data)
	if err != nil {
		http.Error(w, "Server Error", 500)
		log.Println("Template error:", err)
	}
}
