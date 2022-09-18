package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	testConfig := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}

	testStorage := storage.NewMap()
	_ = testStorage.Set("XXX-YYY-ZZZ", "http://ya.ru", &tkn.Token{
		Value:  "xxx",
		Expire: time.Now().Add(tkn.LifeTime),
	})
	_ = testStorage.Set("XXX-YYY-ZZZ", "http://google.com", &tkn.Token{
		Value:  "yyy",
		Expire: time.Now().Add(tkn.LifeTime),
	})

	us2 := &app.URLShortener{
		Config:         testConfig,
		Storage:        testStorage,
		TokenGenerator: generator.NewSimple(),
	}

	type want struct {
		code int
	}
	type request struct {
		method      string
		target      string
		userID      string
		body        string
		contentType string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "success",
			request: request{
				method:      http.MethodDelete,
				target:      "/api/user/urls",
				userID:      "XXX-YYY-ZZZ",
				body:        `["xxx"]`,
				contentType: "application/json; charset=utf-8",
			},
			want: want{
				code: http.StatusAccepted,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, tt.request.target, strings.NewReader(tt.request.body))
			request.Header.Set("Content-type", tt.request.contentType)
			ctx := request.Context()
			ctx = context.WithValue(ctx, middleware.UserIDContextKey, tt.request.userID)
			request = request.WithContext(ctx)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				DeleteURLs(us2, w, r)
			})
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			time.Sleep(100 * time.Millisecond)

			tokens, err := testStorage.GetTokensByUserID(tt.request.userID)
			assert.NoError(t, err)
			var num int
			for _, token := range tokens {
				if token.Removed {
					continue
				}
				num++
			}

			assert.Equal(t, 1, num)
		})
	}
}

func ExampleDeleteURLs() {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"http://localhost:8080/api/user/urls",
		strings.NewReader(`["oTHlXx", "bjHoyQ"]`),
	)
	if err != nil {
		fmt.Println("get error", err)
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	r, err := client.Do(req)
	if err != nil {
		fmt.Println("get error", err)
		return
	}
	defer r.Body.Close()

	fmt.Println(r.StatusCode)
	// 202
}
