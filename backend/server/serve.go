package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"homeserver/config"
	"homeserver/db"

	"golang.org/x/crypto/bcrypt"
)

func SubdomainHandler(sitesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// root domain → serve portfolio
		if host == config.C.RootDomain {
			http.FileServer(http.Dir(sitesDir+"/root")).ServeHTTP(w, r)
			return
		}

		// extract subdomain name
		parts := strings.SplitN(host, ".", 2)
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		name := parts[0]

		// admin subdomain → serve admin dashboard
		if name == "admin" {
			http.FileServer(http.Dir(sitesDir+"/admin")).ServeHTTP(w, r)
			return
		}

		// look up subdomain in DB
		var isPublic bool
		var passwordHash string
		var isActive bool
		err := db.DB.QueryRow(
			"SELECT is_public, COALESCE(password_hash, ''), is_active FROM subdomains WHERE name = ?", name,
		).Scan(&isPublic, &passwordHash, &isActive)

		if err == sql.ErrNoRows || !isActive {
			// fallback to portfolio for unknown hosts
			http.FileServer(http.Dir(sitesDir+"/root")).ServeHTTP(w, r)
			return
		}
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		if !isPublic {
			if !checkPrivateAccess(w, r, passwordHash) {
				return
			}
		}

		http.FileServer(http.Dir(fmt.Sprintf("%s/%s", sitesDir, name))).ServeHTTP(w, r)
	}
}

func checkPrivateAccess(w http.ResponseWriter, r *http.Request, hash string) bool {
	cookie, err := r.Cookie("subdomain_auth")
	if err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(cookie.Value)) == nil {
		return true
	}

	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "subdomain_auth",
				Value:    password,
				Path:     "/",
				HttpOnly: true,
			})
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return false
		}
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, passwordPromptHTML)
	return false
}

const passwordPromptHTML = `<!DOCTYPE html>
<html>
<head><title>Private</title><style>
body{font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#f5f5f5}
form{background:#fff;padding:2rem;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.1);display:flex;flex-direction:column;gap:1rem;min-width:280px}
input{padding:0.5rem;border:1px solid #ddd;border-radius:4px;font-size:1rem}
button{padding:0.5rem;background:#333;color:#fff;border:none;border-radius:4px;cursor:pointer;font-size:1rem}
</style></head>
<body>
<form method="POST">
  <h2 style="margin:0">Private Page</h2>
  <input type="password" name="password" placeholder="Password" autofocus required>
  <button type="submit">Enter</button>
</form>
</body>
</html>`
