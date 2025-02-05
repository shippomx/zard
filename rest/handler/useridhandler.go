package handler

import (
	"context"
	"net/http"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/rest/httpx"
)

func UserIDHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(httpx.KeyUserId) != "" {
			// set user id in context
			newReq := r.WithContext(context.WithValue(r.Context(), logx.UserIDContextKey, r.Header.Get(httpx.KeyUserId)))
			next.ServeHTTP(w, newReq)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
