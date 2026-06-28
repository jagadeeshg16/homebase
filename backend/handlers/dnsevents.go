package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"homeserver/config"
	"homeserver/db"
)

type DNSEvent struct {
	ID        int        `json:"id"`
	Subdomain string     `json:"subdomain"`
	Operation string     `json:"operation"`
	Status    string     `json:"status"`
	ErrorMsg  string     `json:"error_msg"`
	Attempts  int        `json:"attempts"`
	NextRetry *time.Time `json:"next_retry"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func LogDNSEvent(subdomain, operation string) int64 {
	res, err := db.DB.Exec(
		`INSERT INTO dns_events (subdomain, operation, status, attempts) VALUES (?, ?, 'pending', 0)`,
		subdomain, operation,
	)
	if err != nil {
		log.Printf("dns event log failed: %v", err)
		return 0
	}
	id, _ := res.LastInsertId()
	return id
}

func UpdateDNSEvent(id int64, status, errMsg string, attempts int, nextRetry *time.Time) {
	if nextRetry != nil {
		db.DB.Exec(
			`UPDATE dns_events SET status=?, error_msg=?, attempts=?, next_retry=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			status, errMsg, attempts, nextRetry.UTC().Format(time.RFC3339), id,
		)
	} else {
		db.DB.Exec(
			`UPDATE dns_events SET status=?, error_msg=?, attempts=?, next_retry=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			status, errMsg, attempts, id,
		)
	}
}

func GetDNSEvents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, subdomain, operation, status, COALESCE(error_msg,''), attempts,
		       next_retry, created_at, updated_at
		FROM dns_events ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []DNSEvent
	for rows.Next() {
		var e DNSEvent
		var nextRetryStr *string
		rows.Scan(&e.ID, &e.Subdomain, &e.Operation, &e.Status, &e.ErrorMsg,
			&e.Attempts, &nextRetryStr, &e.CreatedAt, &e.UpdatedAt)
		if nextRetryStr != nil {
			t, _ := time.Parse(time.RFC3339, *nextRetryStr)
			e.NextRetry = &t
		}
		events = append(events, e)
	}
	if events == nil {
		events = []DNSEvent{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func RetryDNSEvent(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var subdomain, operation string
	err = db.DB.QueryRow("SELECT subdomain, operation FROM dns_events WHERE id = ?", id).
		Scan(&subdomain, &operation)
	if err != nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	// reset to pending so the retry goroutine picks it up immediately
	db.DB.Exec(
		`UPDATE dns_events SET status='pending', next_retry=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=?`, id,
	)

	go executeDNSEvent(int64(id), subdomain, operation, 0)
	w.WriteHeader(http.StatusAccepted)
}

// StartDNSRetryWorker runs in background, retrying failed/pending events with exponential backoff
func StartDNSRetryWorker() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			retryPending()
		}
	}()
}

func retryPending() {
	rows, err := db.DB.Query(`
		SELECT id, subdomain, operation, attempts FROM dns_events
		WHERE status IN ('pending', 'failed')
		AND (next_retry IS NULL OR next_retry <= datetime('now'))
		LIMIT 10
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	type job struct {
		id        int64
		subdomain string
		operation string
		attempts  int
	}
	var jobs []job
	for rows.Next() {
		var j job
		rows.Scan(&j.id, &j.subdomain, &j.operation, &j.attempts)
		jobs = append(jobs, j)
	}
	rows.Close()

	for _, j := range jobs {
		go executeDNSEvent(j.id, j.subdomain, j.operation, j.attempts)
	}
}

func executeDNSEvent(id int64, subdomain, operation string, prevAttempts int) {
	if DNSProvider == nil || config.C.RootDomain == "" {
		UpdateDNSEvent(id, "failed", "DNS provider not configured", prevAttempts+1, nil)
		return
	}

	attempts := prevAttempts + 1
	var execErr error

	switch operation {
	case "create", "update":
		ip := currentIP()
		if ip == "" {
			execErr = fmt.Errorf("could not fetch public IP")
		} else {
			execErr = DNSProvider.UpsertARecord(config.C.RootDomain, subdomain, ip, 600)
		}
	case "delete":
		execErr = DNSProvider.DeleteRecord(config.C.RootDomain, subdomain)
		// 404 on delete means already gone — treat as success
		if execErr != nil && strings.Contains(execErr.Error(), "404") {
			execErr = nil
		}
	}

	if execErr == nil {
		UpdateDNSEvent(id, "success", "", attempts, nil)
		log.Printf("dns event %d (%s %s): success after %d attempt(s)", id, operation, subdomain, attempts)
		return
	}

	// backoff: 1m, 5m, 15m, 1h, give up after 5
	const maxAttempts = 5
	if attempts >= maxAttempts {
		UpdateDNSEvent(id, "failed", execErr.Error(), attempts, nil)
		log.Printf("dns event %d (%s %s): permanently failed after %d attempts: %v", id, operation, subdomain, attempts, execErr)
		return
	}

	backoffs := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute, 1 * time.Hour}
	delay := backoffs[attempts-1]
	if attempts-1 >= len(backoffs) {
		delay = time.Hour
	}
	next := time.Now().Add(delay)
	UpdateDNSEvent(id, "failed", execErr.Error(), attempts, &next)
	log.Printf("dns event %d (%s %s): failed (attempt %d/%d), retry in %s: %v",
		id, operation, subdomain, attempts, maxAttempts, delay, execErr)
}
