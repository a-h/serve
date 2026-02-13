package handlers

import (
	"crypto/subtle"
	"net/http"
)

func NewBasicAuthMiddleware(next http.Handler, username, password string) http.Handler {
	return &BasicAuthMiddleware{
		next:     next,
		username: username,
		password: password,
	}
}

type BasicAuthMiddleware struct {
	next     http.Handler
	username string
	password string
}

func (m *BasicAuthMiddleware) credentialsMatch(user, pass string) bool {
	userMatches := subtle.ConstantTimeCompare([]byte(user), []byte(m.username)) == 1
	passMatches := subtle.ConstantTimeCompare([]byte(pass), []byte(m.password)) == 1
	return userMatches && passMatches
}

func (m *BasicAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || !m.credentialsMatch(user, pass) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	m.next.ServeHTTP(w, r)
}
