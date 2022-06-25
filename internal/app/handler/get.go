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

func Get(us Getter, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

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

		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
	_, err = w.Write([]byte(url))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func GetUserURLs(us Getter, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

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

		http.Error(w, err.Error(), 500)
		return
	}

	result, err := json.Marshal(urls)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = w.Write(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
