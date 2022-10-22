package handler

import (
	"context"
	"net/http"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

// Ping checks the database connection.
func Ping(us Pinger) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if err := us.Ping(r.Context()); err != nil {
			http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
			return
		}
	}
	return fn
}
