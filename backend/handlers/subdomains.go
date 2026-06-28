package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"homeserver/config"
	"homeserver/db"
	dnspkg "homeserver/dns"

	"golang.org/x/crypto/bcrypt"
)

var DNSProvider dnspkg.Provider

type Subdomain struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	FullDomain string    `json:"full_domain"`
	Type       string    `json:"type"`
	ProxyURL   string    `json:"proxy_url"`
	IsPublic   bool      `json:"is_public"`
	IsActive   bool      `json:"is_active"`
	RateLimit  int       `json:"rate_limit"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ListSubdomains(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`SELECT id, name, full_domain, COALESCE(type,'static'), COALESCE(proxy_url,''), is_public, rate_limit, is_active, created_at, updated_at FROM subdomains ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subs []Subdomain
	for rows.Next() {
		var s Subdomain
		rows.Scan(&s.ID, &s.Name, &s.FullDomain, &s.Type, &s.ProxyURL, &s.IsPublic, &s.RateLimit, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
		subs = append(subs, s)
	}
	if subs == nil {
		subs = []Subdomain{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

func CreateSubdomain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		Type      string `json:"type"`
		ProxyURL  string `json:"proxy_url"`
		IsPublic  bool   `json:"is_public"`
		RateLimit int    `json:"rate_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	req.Name = strings.ToLower(strings.TrimSpace(req.Name))
	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = "static"
	}
	fullDomain := req.Name + "." + config.C.RootDomain

	// insert as active (manual creation by admin = intentional)
	_, err := db.DB.Exec(
		`INSERT INTO subdomains (name, full_domain, type, proxy_url, is_public, is_active, rate_limit) VALUES (?, ?, ?, ?, ?, 1, ?)`,
		req.Name, fullDomain, req.Type, req.ProxyURL, req.IsPublic, req.RateLimit,
	)
	if err != nil {
		http.Error(w, "subdomain already exists: "+err.Error(), http.StatusConflict)
		return
	}

	// create folder + placeholder
	sitesDir := fmt.Sprintf("/home/%s/server/sites/%s", os.Getenv("USER"), req.Name)
	if err := os.MkdirAll(sitesDir, 0755); err != nil {
		log.Printf("create folder failed for %s: %v", req.Name, err)
	} else {
		placeholder := fmt.Sprintf(`<!DOCTYPE html><html><head><title>%s</title></head><body><h1>%s</h1><p>Drop your files here.</p></body></html>`, req.Name, fullDomain)
		os.WriteFile(sitesDir+"/index.html", []byte(placeholder), 0644)
	}

	// create DNS A record via tracked event
	if DNSProvider != nil && config.C.RootDomain != "" {
		eventID := LogDNSEvent(req.Name, "create")
		go executeDNSEvent(eventID, req.Name, "create", 0)
	}

	w.WriteHeader(http.StatusCreated)
}

func DeleteSubdomain(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/subdomains/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var name string
	db.DB.QueryRow("SELECT name FROM subdomains WHERE id = ?", id).Scan(&name)

	db.DB.Exec("DELETE FROM subdomains WHERE id = ?", id)

	// delete DNS record via tracked event
	if DNSProvider != nil && name != "" && config.C.RootDomain != "" {
		eventID := LogDNSEvent(name, "delete")
		go executeDNSEvent(eventID, name, "delete", 0)
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		IsPublic bool   `json:"is_public"`
		IsActive *bool  `json:"is_active"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// fetch current state
	var name string
	var wasActive bool
	db.DB.QueryRow("SELECT name, is_active FROM subdomains WHERE id = ?", id).Scan(&name, &wasActive)

	var hash string
	if !req.IsPublic && req.Password != "" {
		b, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		hash = string(b)
	}

	if req.IsActive != nil {
		db.DB.Exec(
			"UPDATE subdomains SET is_public = ?, is_active = ?, password_hash = CASE WHEN ? != '' THEN ? ELSE password_hash END, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			req.IsPublic, *req.IsActive, hash, hash, id,
		)
		// if activating for the first time, create DNS record via tracked event
		if *req.IsActive && !wasActive && DNSProvider != nil && config.C.RootDomain != "" {
			eventID := LogDNSEvent(name, "create")
			go executeDNSEvent(eventID, name, "create", 0)
		}
	} else {
		db.DB.Exec(
			"UPDATE subdomains SET is_public = ?, password_hash = CASE WHEN ? != '' THEN ? ELSE password_hash END, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			req.IsPublic, hash, hash, id,
		)
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		RateLimit int `json:"rate_limit"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	db.DB.Exec(
		"UPDATE subdomains SET rate_limit = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.RateLimit, id,
	)
	w.WriteHeader(http.StatusOK)
}

func currentIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	return strings.TrimSpace(string(buf[:n]))
}
