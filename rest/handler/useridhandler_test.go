package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/rest/httpx"
)

func TestUserIDHandler(t *testing.T) {
	t.Run("UserIDPresent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(httpx.KeyUserId, "12345")

		handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(logx.UserIDContextKey)
			if userID != "12345" {
				t.Errorf("UserID not set in context, expected: 12345, got: %v", userID)
			}
		})

		UserIDHandler(handler).ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("UserIDNotPresent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(logx.UserIDContextKey)
			if userID != nil {
				t.Errorf("UserID should not be set in context when KeyUserId is not present, got: %v", userID)
			}
		})

		UserIDHandler(handler).ServeHTTP(httptest.NewRecorder(), req)
	})
}
