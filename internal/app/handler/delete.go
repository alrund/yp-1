package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/alrund/yp-1/internal/app/middleware"
)

// DeleteURLs deletes shortened URL tokens.
func (hc *Collection) DeleteURLs() func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
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

		var tokens []string
		err = json.Unmarshal(b, &tokens)
		if err != nil {
			http.Error(w, "400 Bad Request.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)

		go func() {
			_ = hc.us.RemoveTokens(tokens, userID)
		}()
	}
	return fn
}
