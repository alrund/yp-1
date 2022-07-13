package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
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

	jsonRequests := make([]JSONBatchRequestRow, 0)
	err = json.Unmarshal(b, &jsonRequests)
	if err != nil {
		http.Error(w, "400 Bad Request.", http.StatusBadRequest)
		return
	}

	jsonResponse := make([]JSONBatchResponseRow, 0)
	URLs, URL2Row := getURL2Row(jsonRequests)
	tokens, err := us.AddBatch(userID, URLs)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpCode = http.StatusConflict
	}

	for URL, token := range tokens {
		row, ok := URL2Row[URL]
		if !ok {
			http.Error(w, "URL not found in URL2Row map", http.StatusInternalServerError)
			return
		}
		correlationID := row.CorrelationID
		jsonResponse = append(jsonResponse, JSONBatchResponseRow{
			CorrelationID: correlationID,
			ShortURL:      us.GetBaseURL() + token.Value,
		})
	}

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

func getURL2Row(rows []JSONBatchRequestRow) ([]string, map[string]JSONBatchRequestRow) {
	URLs := make([]string, 0)
	URL2Row := map[string]JSONBatchRequestRow{}

	for _, row := range rows {
		URLs = append(URLs, row.OriginalURL)
		URL2Row[row.OriginalURL] = row
	}

	return URLs, URL2Row
}
