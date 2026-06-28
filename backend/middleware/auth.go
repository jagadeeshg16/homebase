package middleware

import (
	"context"
	"net/http"

	"homeserver/config"
)

type contextKey string

const sessionKey contextKey = "authenticated"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != config.C.SessionSecret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func InternalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Internal-Token")
		if token == "" || token != config.C.InternalToken {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
