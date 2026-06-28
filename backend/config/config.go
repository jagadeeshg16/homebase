package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	InternalToken string
	SessionSecret string

	DNSProvider string
	RootDomain  string

	GoDaddyAPIKey    string
	GoDaddyAPISecret string

	CloudflareAPIToken string
	CloudflareZoneID   string

	DBPath string
}

var C Config

func Load() {
	if err := godotenv.Load("/home/" + os.Getenv("USER") + "/server/.env"); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	C = Config{
		Port:               getenv("PORT", "8080"),
		InternalToken:      getenv("INTERNAL_TOKEN", ""),
		SessionSecret:      getenv("SESSION_SECRET", "change-me"),
		DNSProvider:        getenv("DNS_PROVIDER", "godaddy"),
		RootDomain:         getenv("ROOT_DOMAIN", ""),
		GoDaddyAPIKey:      getenv("GODADDY_API_KEY", ""),
		GoDaddyAPISecret:   getenv("GODADDY_API_SECRET", ""),
		CloudflareAPIToken: getenv("CLOUDFLARE_API_TOKEN", ""),
		CloudflareZoneID:   getenv("CLOUDFLARE_ZONE_ID", ""),
		DBPath:             getenv("DB_PATH", "/home/"+os.Getenv("USER")+"/server/data/server.db"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
