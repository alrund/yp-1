package app

import (
	"errors"
	stg "github.com/alrund/yp-1/internal/app/storage"
	"io"
	"net/http"
)

var us = &URLShortener{stg.NewMapStorage()}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if r.URL.Path != "/" {
			http.Error(w, "400 Bad Request.", http.StatusBadRequest)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		token, err := us.Add(string(b))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(token.Value))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		return

	case http.MethodGet:
		id := r.URL.Path[1:]
		if id == "" {
			http.Error(w, "400 Bad Request.", http.StatusBadRequest)
			return
		}
		url, err := us.Get(id)
		if err != nil {
			if errors.Is(err, stg.ErrTokenNotFound) {
				http.Error(w, "404 Not Found.", http.StatusNotFound)
				return
			}

			if errors.Is(err, ErrTokenExpiredError) {
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

	default:
		http.Error(w, "Only GET & POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}
