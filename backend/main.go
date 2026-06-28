package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"homeserver/config"
	"homeserver/db"
	dnspkg "homeserver/dns"
	"homeserver/handlers"
	"homeserver/middleware"
	"homeserver/server"
)

func main() {
	config.Load()
	db.Init(config.C.DBPath)

	provider := newDNSProvider()
	handlers.DNSProvider = provider

	sitesDir := fmt.Sprintf("/home/%s/server/sites", os.Getenv("USER"))

	mux := http.NewServeMux()

	// Public auth routes
	mux.HandleFunc("/api/auth/login", handlers.Login)
	mux.HandleFunc("/api/auth/logout", handlers.Logout)

	// DDNS update — internal token only (called by ddns.sh cron)
	mux.Handle("/api/dns/update", middleware.InternalOnly(http.HandlerFunc(handlers.UpdateDNS)))

	// All other API routes — require session cookie
	protected := http.NewServeMux()

	protected.HandleFunc("/api/subdomains", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListSubdomains(w, r)
		case http.MethodPost:
			handlers.CreateSubdomain(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	protected.HandleFunc("/api/subdomains/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/privacy") && r.Method == http.MethodPatch:
			handlers.UpdatePrivacy(w, r)
		case strings.HasSuffix(path, "/ratelimit") && r.Method == http.MethodPatch:
			handlers.UpdateRateLimit(w, r)
		case r.Method == http.MethodDelete:
			handlers.DeleteSubdomain(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	protected.HandleFunc("/api/health", handlers.GetHealth)
	protected.HandleFunc("/api/health/", handlers.GetSubdomainHealth)
	protected.HandleFunc("/api/settings/password", handlers.ChangePassword)

	mux.Handle("/api/", middleware.Auth(protected))

	// Subdomain static file serving with privacy gate
	mux.HandleFunc("/", server.SubdomainHandler(sitesDir))

	// Watch sites/ dir — auto-register new folders as private+inactive
	server.WatchSites(sitesDir)

	// Health check goroutine — runs every 60s
	go func() {
		for {
			time.Sleep(60 * time.Second)
			handlers.RunHealthChecks()
		}
	}()

	addr := ":" + config.C.Port
	log.Println("server starting on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func newDNSProvider() dnspkg.Provider {
	switch config.C.DNSProvider {
	case "cloudflare":
		return &dnspkg.Cloudflare{
			APIToken: config.C.CloudflareAPIToken,
			ZoneID:   config.C.CloudflareZoneID,
		}
	default:
		return &dnspkg.GoDaddy{
			APIKey:    config.C.GoDaddyAPIKey,
			APISecret: config.C.GoDaddyAPISecret,
		}
	}
}
