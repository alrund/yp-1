package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"mime"
	"net/http"

	tkn "github.com/alrund/yp-1/internal/app/token"
)

type Adder interface {
	GetBaseURL() string
	Add(userID, url string) (*tkn.Token, error)
}

type JSONRequest struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

func Add(us Adder, w http.ResponseWriter, r *http.Request) {
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

	userID := getUserId()

	token, err := us.Add(userID, string(b))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(us.GetBaseURL() + token.Value))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func AddJSON(us Adder, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/api/shorten" {
		http.Error(w, "400 Bad Request.", http.StatusBadRequest)
		return
	}

	if !hasContentType(r, "application/json") {
		http.Error(w, "415 Unsupported Media Type.", http.StatusUnsupportedMediaType)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jsonRequest := JSONRequest{}
	err = json.Unmarshal(b, &jsonRequest)
	if err != nil {
		http.Error(w, "400 Bad Request.", http.StatusBadRequest)
		return
	}

	userID := getUserId()

	token, err := us.Add(userID, jsonRequest.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jsonResponse := JSONResponse{Result: us.GetBaseURL() + token.Value}
	result, err := json.Marshal(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
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

func getUserId() string {
	id := uuid.New()
	return id.String()
}
