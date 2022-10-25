package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
)

type JSONRequest struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

// Add adds a URL string to shorten.
func (hc *Collection) Add() func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
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

		token, err := hc.us.Add(userID, string(b))
		if err != nil {
			if !errors.Is(err, storage.ErrURLAlreadyExists) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			httpCode = http.StatusConflict
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(httpCode)
		_, err = w.Write([]byte(hc.us.GetBaseURL() + token.Value))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return fn
}

// AddJSON adds a URL string to shorten as a JSON object.
func (hc *Collection) AddJSON() func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
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

		token, err := hc.us.Add(userID, jsonRequest.URL)
		if err != nil {
			if !errors.Is(err, storage.ErrURLAlreadyExists) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			httpCode = http.StatusConflict
		}

		jsonResponse := JSONResponse{Result: hc.us.GetBaseURL() + token.Value}
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
	return fn
}

func hasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	t, _, err := mime.ParseMediaType(contentType)
	if err == nil && t == mimetype {
		return true
	}
	return false
}
