package handler

import (
	"errors"
	"net/http"

	stg "github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
)

type Getter interface {
	Get(tokenValue string) (string, error)
}

func GetHandler(us Getter, w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, stg.ErrTokenNotFound) {
			http.Error(w, "404 Not Found.", http.StatusNotFound)
			return
		}

		if errors.Is(err, tkn.ErrTokenExpiredError) {
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
