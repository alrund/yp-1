package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const UserIDContextKey ContextKey = "userID"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), UserIDContextKey, uuid.New().String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
