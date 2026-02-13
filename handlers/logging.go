package handlers

import (
	"log/slog"
	"net/http"
	"time"
)

func NewLoggingMiddleware(log *slog.Logger, logRemoteAddr bool, next http.Handler) http.Handler {
	return &LoggingMiddleware{
		log:           log,
		logRemoteAddr: logRemoteAddr,
		next:          next,
	}
}

type LoggingMiddleware struct {
	log           *slog.Logger
	logRemoteAddr bool
	next          http.Handler
}

func (m *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	args := []any{
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
	}
	if m.logRemoteAddr {
		args = append(args, slog.String("remote_addr", r.RemoteAddr))
	}

	start := time.Now()

	sw := &statusWriter{ResponseWriter: w}
	m.next.ServeHTTP(sw, r)
	args = append(args, slog.Duration("duration", time.Since(start)), slog.Int("status", sw.Status()))

	m.log.Info("Request", args...)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
