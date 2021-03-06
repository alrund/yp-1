package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
)

type Adder interface {
	GetBaseURL() string
	Add(userID, url string) (*tkn.Token, error)
	AddBatch(userID string, urls []string) (map[string]*tkn.Token, error)
}

type JSONRequest struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

func Add(us Adder, w http.ResponseWriter, r *http.Request) {
	httpCode := http.StatusCreated

	contextUserID := r.Context().Value(middleware.UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := us.Add(userID, string(b))
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpCode = http.StatusConflict
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(httpCode)
	_, err = w.Write([]byte(us.GetBaseURL() + token.Value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AddJSON(us Adder, w http.ResponseWriter, r *http.Request) {
	httpCode := http.StatusCreated

	if !hasContentType(r, "application/json") {
		http.Error(w, "415 Unsupported Media Type.", http.StatusUnsupportedMediaType)
		return
	}

	contextUserID := r.Context().Value(middleware.UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonRequest := JSONRequest{}
	err = json.Unmarshal(b, &jsonRequest)
	if err != nil {
		http.Error(w, "400 Bad Request.", http.StatusBadRequest)
		return
	}

	token, err := us.Add(userID, jsonRequest.URL)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpCode = http.StatusConflict
	}

	jsonResponse := JSONResponse{Result: us.GetBaseURL() + token.Value}
	result, err := json.Marshal(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpCode)
	_, err = w.Write(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func hasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	t, _, err := mime.ParseMediaType(contentType)
	if err == nil && t == mimetype {
		return true
	}
	return false
}
