package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuthMiddleware(t *testing.T) {
	var handlerCalled bool
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		handlerCalled = true
	})
	authMiddleware := NewBasicAuthMiddleware(protected, "admin", "secret")
	t.Run("no credentials return 401", func(t *testing.T) {
		handlerCalled = false

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		authMiddleware.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})
	t.Run("invalid credentials return 401", func(t *testing.T) {
		handlerCalled = false

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetBasicAuth("admin", "wrongpassword")
		w := httptest.NewRecorder()
		authMiddleware.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})
	t.Run("valid credentials call next handler", func(t *testing.T) {
		handlerCalled = false

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetBasicAuth("admin", "secret")
		w := httptest.NewRecorder()
		authMiddleware.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if !handlerCalled {
			t.Error("expected next handler to be called, but it was not")
		}
	})
}
