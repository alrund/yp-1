package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alrund/yp-1/internal/app/encryption"
	"github.com/google/uuid"
)

type ContextKey string

const UserIDContextKey ContextKey = "userID"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetCookie(r)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
			return
		}

		if userID == "" {
			userID = uuid.New().String()
			err = AddCookie(userID, w)
			if err != nil {
				http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
				return
			}
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AddCookie(userID string, w http.ResponseWriter) error {
	encrypted, err := encryption.Encrypt(userID)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     string(UserIDContextKey),
		Value:    encrypted,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
	})

	return nil
}

func GetCookie(r *http.Request) (string, error) {
	userCookie, err := r.Cookie(string(UserIDContextKey))
	if err != nil {
		return "", err
	}

	if userCookie.Value == "" {
		return "", nil
	}

	userID, err := encryption.Decrypt(userCookie.Value)
	if err != nil {
		return "", err
	}

	return userID, nil
}
