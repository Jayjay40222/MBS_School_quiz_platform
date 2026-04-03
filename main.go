package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	InitDB()
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/login/", handleLogin)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/signup/", handleSignup)
	mux.HandleFunc("/signup", handleSignup)
	mux.HandleFunc("/dashboard", handleDashboard)
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

			if err := bcrypt.CompareHashAndPassword([]byte(usr.hashedPassword), []byte(password)); err != nil || usr.grade != class {
				w.Write([]byte("Incorrect password or class"))
				return
			}

			// Generate session token
			sessionToken := generateSessionToken()
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    sessionToken,
				HttpOnly: true,
				Secure:   false,
				Path:     "/",
			})

			err = AddSessionToDB(usr.id, sessionToken)
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

func generateSessionToken() string {
	b := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // handle properly in real apps
	}
	return base64.URLEncoding.EncodeToString(b)
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
			fullName:       username,
			hashedPassword: string(hashPass),
			grade:          class,
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

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Println("No session cookie found. Redirecting to /login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	usr, err := GetUserFromSession(cookie.Value)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if usr == nil {
		log.Println("Invalid session. Redirecting to /login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// User is authenticated — serve dashboard
	http.ServeFile(w, r, "templates/dashboard.html")
}
