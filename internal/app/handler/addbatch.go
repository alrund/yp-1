package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alrund/yp-1/internal/app/middleware"
)

type JSONBatchRequestRow struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type JSONBatchResponseRow struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func AddBatchJSON(us Adder, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/api/shorten/batch" {
		http.Error(w, "400 Bad Request.", http.StatusBadRequest)
		return
	}

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
		http.Error(w, err.Error(), 500)
		return
	}

	jsonRequests := make([]JSONBatchRequestRow, 0)
	err = json.Unmarshal(b, &jsonRequests)
	if err != nil {
		http.Error(w, "400 Bad Request.", http.StatusBadRequest)
		return
	}

	jsonResponse := make([]JSONBatchResponseRow, 0)
	for _, jsonRequest := range jsonRequests {
		token, err := us.Add(userID, jsonRequest.OriginalURL)
		fmt.Println(userID)
		fmt.Println(jsonRequest.OriginalURL)

		if err != nil {
			fmt.Println(err.Error())

			http.Error(w, err.Error(), 500)
			return
		}

		jsonResponse = append(jsonResponse, JSONBatchResponseRow{
			CorrelationID: jsonRequest.CorrelationID,
			ShortURL:      us.GetBaseURL() + token.Value,
		})
	}

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
