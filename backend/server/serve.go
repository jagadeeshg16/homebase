package server

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"homeserver/config"
	"homeserver/db"
)

type subdomainRecord struct {
	IsPublic     bool
	IsActive     bool
	Type         string
	ProxyURL     string
	PasswordHash string
}

func SubdomainHandler(sitesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// root domain → portfolio
		if host == config.C.RootDomain || host == "" {
			http.FileServer(http.Dir(sitesDir+"/root")).ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(host, ".", 2)
		if len(parts) < 2 {
			http.FileServer(http.Dir(sitesDir+"/root")).ServeHTTP(w, r)
			return
		}
		name := parts[0]

		// admin → React SPA
		if name == "admin" {
			serveReactApp(sitesDir+"/admin", w, r)
			return
		}

		// look up subdomain
		var rec subdomainRecord
		err := db.DB.QueryRow(`
			SELECT is_public, is_active, COALESCE(type,'static'), COALESCE(proxy_url,''), COALESCE(password_hash,'')
			FROM subdomains WHERE name = ?`, name,
		).Scan(&rec.IsPublic, &rec.IsActive, &rec.Type, &rec.ProxyURL, &rec.PasswordHash)

		if err == sql.ErrNoRows || !rec.IsActive {
			http.FileServer(http.Dir(sitesDir+"/root")).ServeHTTP(w, r)
			return
		}
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		if !rec.IsPublic {
			// admin session → bypass private check
			if !isAdminSession(r) {
				http.NotFound(w, r)
				return
			}
		}

		switch rec.Type {
		case "proxy":
			// browser implicit form submit (password field + button) sends POST /
			// redirect to GET so the upstream app receives a normal page load
			if r.Method == http.MethodPost && r.URL.Path == "/" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			proxyTo(rec.ProxyURL, w, r)
		default:
			http.FileServer(http.Dir(fmt.Sprintf("%s/%s", sitesDir, name))).ServeHTTP(w, r)
		}
	}
}

func proxyTo(target string, w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(target)
	if err != nil {
		http.Error(w, "invalid proxy target", http.StatusInternalServerError)
		return
	}

	// WebSocket — raw TCP tunnel
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		proxyWebSocket(u, w, r)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error for %s: %v", target, err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	outReq := r.Clone(r.Context())
	outReq.Host = u.Host
	proxy.ServeHTTP(w, outReq)
}

func proxyWebSocket(target *url.URL, w http.ResponseWriter, r *http.Request) {
	addr := target.Host
	if target.Port() == "" {
		if target.Scheme == "https" {
			addr += ":443"
		} else {
			addr += ":80"
		}
	}

	upstream, err := net.Dial("tcp", addr)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer upstream.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket not supported", http.StatusInternalServerError)
		return
	}
	client, _, err := hj.Hijack()
	if err != nil {
		return
	}
	defer client.Close()

	// forward original HTTP upgrade request to upstream
	r.Write(upstream)

	// bidirectional pipe
	done := make(chan struct{}, 2)
	go func() { io.Copy(upstream, client); done <- struct{}{} }()
	go func() { io.Copy(client, upstream); done <- struct{}{} }()
	<-done
}

func isAdminSession(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	return err == nil && cookie.Value == config.C.SessionSecret
}

func serveReactApp(dir string, w http.ResponseWriter, r *http.Request) {
	fs := http.Dir(dir)
	if r.URL.Path != "/" {
		f, err := fs.Open(r.URL.Path)
		if err == nil {
			f.Close()
			http.FileServer(fs).ServeHTTP(w, r)
			return
		}
	}
	http.ServeFile(w, r, dir+"/index.html")
}

