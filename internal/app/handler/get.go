package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token"
)

type Getter interface {
	Get(tokenValue string) (string, error)
	GetUserURLs(userID string) ([]storage.URLpairs, error)
}

// Get returns a URL by token.
func Get(us Getter) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[1:]
		if id == "" {
			http.Error(w, "400 Bad Request.", http.StatusBadRequest)
			return
		}
		url, err := us.Get(id)
		if err != nil {
			if errors.Is(err, storage.ErrTokenNotFound) {
				http.Error(w, "404 Not Found.", http.StatusNotFound)
				return
			}

			if errors.Is(err, token.ErrTokenExpiredError) {
				http.Error(w, "498 Invalid Token.", 498)
				return
			}

			if errors.Is(err, token.ErrTokenRemovedError) {
				http.Error(w, "410 Gone.", http.StatusGone)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		_, err = w.Write([]byte(url))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return fn
}

// GetUserURLs returns a URL by user ID.
func GetUserURLs(us Getter) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		contextUserID := r.Context().Value(middleware.UserIDContextKey)
		userID, ok := contextUserID.(string)
		if !ok {
			http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
			return
		}

		urls, err := us.GetUserURLs(userID)
		if err != nil {
			if errors.Is(err, storage.ErrTokenNotFound) {
				http.Error(w, "204 No Content.", http.StatusNoContent)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result, err := json.Marshal(urls)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err = w.Write(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return fn
}
