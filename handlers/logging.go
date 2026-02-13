package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func NewLoggingMiddleware(logRemoteAddr bool, next http.Handler) http.Handler {
	return &LoggingMiddleware{
		logRemoteAddr: logRemoteAddr,
		next:          next,
	}
}

type LoggingMiddleware struct {
	logRemoteAddr bool
	next          http.Handler
}

func (m *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sb strings.Builder
	sb.WriteString(time.Now().Format(time.RFC3339))
	sb.WriteString(" ")
	sb.WriteString(r.Method)
	sb.WriteString(" ")
	sb.WriteString(r.URL.String())
	if m.logRemoteAddr {
		sb.WriteString(" ")
		sb.WriteString(r.RemoteAddr)
	}
	fmt.Println(sb.String())

	m.next.ServeHTTP(w, r)
}
