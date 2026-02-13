package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

func Create(dir string, logRemoteAddr bool, isWritable bool, auth string) (h http.Handler, closer func() error, err error) {
	handler, closer, err := NewFileHandler(dir, isWritable)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create file handler: %w", err)
	}
	withLogging := NewLoggingMiddleware(logRemoteAddr, handler)
	if auth != "" {
		parts := strings.SplitN(auth, ":", 2)
		if len(parts) != 2 {
			return nil, closer, fmt.Errorf("-auth must be in the format username:password")
		}
		withAuth := NewBasicAuthMiddleware(withLogging, parts[0], parts[1])
		return withAuth, closer, nil
	}
	return withLogging, closer, nil
}
