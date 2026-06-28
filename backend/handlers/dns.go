package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"homeserver/config"
	"homeserver/db"

	"golang.org/x/crypto/bcrypt"
)

func UpdateDNS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IP == "" {
		http.Error(w, "ip required", http.StatusBadRequest)
		return
	}

	// root + wildcard covers everything; also update individual records for explicitness
	targets := map[string]bool{"@": true, "*": true}
	rows, err := db.DB.Query("SELECT name FROM subdomains WHERE is_active = 1")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			rows.Scan(&name)
			targets[name] = true
		}
	}

	success := true
	for name := range targets {
		if err := DNSProvider.UpsertARecord(config.C.RootDomain, name, req.IP, 600); err != nil {
			log.Printf("dns update failed for %s: %v", name, err)
			success = false
		}
	}

	db.DB.Exec("INSERT INTO dns_log (new_ip, provider, success) VALUES (?, ?, ?)",
		req.IP, config.C.DNSProvider, success)

	if !success {
		http.Error(w, "some DNS updates failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(req.NewPassword) < 8 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}

	var username, hash string
	if err := db.DB.QueryRow("SELECT username, password_hash FROM users LIMIT 1").Scan(&username, &hash); err != nil {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "incorrect current password", http.StatusUnauthorized)
		return
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	db.DB.Exec("UPDATE users SET password_hash = ? WHERE username = ?", string(newHash), username)
	w.WriteHeader(http.StatusOK)
}
