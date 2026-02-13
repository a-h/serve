package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/serve/config"
)

func Create(log *slog.Logger, conf *config.Config) (h http.Handler, closer func() error, err error) {
	handler, closer, err := NewFileHandler(log, conf.Dir, conf.ReadOnly)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create file handler: %w", err)
	}
	withLogging := NewLoggingMiddleware(log, conf.LogRemoteAddr, handler)
	if conf.Auth != "" {
		parts := strings.SplitN(conf.Auth, ":", 2)
		if len(parts) != 2 {
			return nil, closer, fmt.Errorf("-auth must be in the format username:password")
		}
		withAuth := NewBasicAuthMiddleware(withLogging, parts[0], parts[1])
		return withAuth, closer, nil
	}
	return withLogging, closer, nil
}
