package handler

import (
	"net/http"
)

// Ping checks the database connection.
func (hc *Collection) Ping() func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if err := hc.us.Ping(r.Context()); err != nil {
			http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
			return
		}
	}
	return fn
}
