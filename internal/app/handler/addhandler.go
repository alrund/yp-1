package handler

import (
	"io"
	"net/http"

	tkn "github.com/alrund/yp-1/internal/app/token"
)

type Adder interface {
	GetServerURL() string
	Add(url string) (*tkn.Token, error)
}

func AddHandler(us Adder, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(us.GetServerURL() + token.Value))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
