package server

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"homeserver/config"
	"homeserver/db"

	"github.com/fsnotify/fsnotify"
)

func WatchSites(sitesDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("watcher init:", err)
	}

	if err := watcher.Add(sitesDir); err != nil {
		log.Fatal("watcher add:", err)
	}

	log.Println("watching for new subdomains in", sitesDir)

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == 0 {
					continue
				}
				info, err := os.Stat(event.Name)
				if err != nil || !info.IsDir() {
					continue
				}
				name := filepath.Base(event.Name)
				// skip hidden dirs and reserved names
				if strings.HasPrefix(name, ".") || name == "root" || name == "admin" {
					continue
				}
				registerSubdomain(name)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()
}

func registerSubdomain(name string) {
	fullDomain := name + "." + config.C.RootDomain

	var exists int
	db.DB.QueryRow("SELECT COUNT(*) FROM subdomains WHERE name = ?", name).Scan(&exists)
	if exists > 0 {
		return
	}

	_, err := db.DB.Exec(
		`INSERT INTO subdomains (name, full_domain, is_public, is_active, rate_limit) VALUES (?, ?, 0, 0, 100)`,
		name, fullDomain,
	)
	if err != nil {
		log.Printf("auto-register failed for %s: %v", name, err)
		return
	}
	log.Printf("auto-registered subdomain: %s (private, inactive — activate via admin)", name)
}
