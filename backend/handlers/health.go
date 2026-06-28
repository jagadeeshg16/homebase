package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"homeserver/db"
)

type HealthEntry struct {
	SubdomainID   int       `json:"subdomain_id"`
	SubdomainName string    `json:"subdomain_name"`
	Status        int       `json:"status"`
	ResponseMs    int       `json:"response_ms"`
	CheckedAt     time.Time `json:"checked_at"`
}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT h.subdomain_id, s.name, h.status, h.response_ms, h.checked_at
		FROM health_logs h
		JOIN subdomains s ON s.id = h.subdomain_id
		WHERE h.id IN (
			SELECT MAX(id) FROM health_logs GROUP BY subdomain_id
		)
		ORDER BY s.name
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []HealthEntry
	for rows.Next() {
		var e HealthEntry
		rows.Scan(&e.SubdomainID, &e.SubdomainName, &e.Status, &e.ResponseMs, &e.CheckedAt)
		entries = append(entries, e)
	}
	json.NewEncoder(w).Encode(entries)
}

func GetSubdomainHealth(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query(`
		SELECT subdomain_id, status, response_ms, checked_at
		FROM health_logs WHERE subdomain_id = ?
		ORDER BY checked_at DESC LIMIT 50
	`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []HealthEntry
	for rows.Next() {
		var e HealthEntry
		rows.Scan(&e.SubdomainID, &e.Status, &e.ResponseMs, &e.CheckedAt)
		entries = append(entries, e)
	}
	json.NewEncoder(w).Encode(entries)
}

func RunHealthChecks() {
	rows, err := db.DB.Query("SELECT id, full_domain, COALESCE(type,'static'), COALESCE(proxy_url,'') FROM subdomains WHERE is_active = 1")
	if err != nil {
		return
	}
	defer rows.Close()

	type sub struct {
		ID       int
		Domain   string
		Type     string
		ProxyURL string
	}
	var subs []sub
	for rows.Next() {
		var s sub
		rows.Scan(&s.ID, &s.Domain, &s.Type, &s.ProxyURL)
		subs = append(subs, s)
	}

	for _, s := range subs {
		target := fmt.Sprintf("http://%s", s.Domain)
		if s.Type == "proxy" && s.ProxyURL != "" {
			target = s.ProxyURL
		}
		start := time.Now()
		resp, err := http.Get(target)
		elapsed := int(time.Since(start).Milliseconds())
		status := 0
		if err == nil {
			status = resp.StatusCode
			resp.Body.Close()
		}
		db.DB.Exec(
			"INSERT INTO health_logs (subdomain_id, status, response_ms) VALUES (?, ?, ?)",
			s.ID, status, elapsed,
		)
	}
}
