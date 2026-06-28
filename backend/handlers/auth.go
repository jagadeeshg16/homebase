package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"homeserver/config"
	"homeserver/db"

	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var hash string
	err := db.DB.QueryRow("SELECT password_hash FROM users WHERE username = ?", req.Username).Scan(&hash)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    config.C.SessionSecret,
		Path:     "/",
		Domain:   "." + config.C.RootDomain,
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	w.WriteHeader(http.StatusOK)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
	w.WriteHeader(http.StatusOK)
}
