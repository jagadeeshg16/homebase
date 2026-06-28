package db

import (
	"database/sql"
	_ "embed"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

var DB *sql.DB

func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if _, err = DB.Exec(schema); err != nil {
		log.Fatalf("db migrate: %v", err)
	}
	migrate()
	log.Println("database ready:", path)
}

// migrate adds new columns to existing installs safely
func migrate() {
	migrations := []string{
		`ALTER TABLE subdomains ADD COLUMN type TEXT DEFAULT 'static'`,
		`ALTER TABLE subdomains ADD COLUMN proxy_url TEXT`,
	}
	for _, m := range migrations {
		// sqlite returns error if column already exists — ignore it
		DB.Exec(m)
	}
}
