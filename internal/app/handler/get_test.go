package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/helper"
	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/stretchr/testify/assert"
)

type Storage struct{}

func (st *Storage) HasToken(tokenValue string) (bool, error) {
	switch tokenValue {
	case "qwerty":
		return true, nil
	case "expired":
		return true, nil
	case "removed":
		return true, nil
	}
	return false, nil
}

func (st *Storage) GetToken(tokenValue string) (*tkn.Token, error) {
	switch tokenValue {
	case "expired":
		return &tkn.Token{Value: "expired", Expire: time.Now().Add(-tkn.LifeTime)}, nil
	case "removed":
		return &tkn.Token{Value: "removed", Expire: time.Now().Add(tkn.LifeTime), Removed: true}, nil
	default:
		return &tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)}, nil
	}
}

func (st *Storage) GetURLsByUserID(userID string) ([]storage.URLpairs, error) {
	switch userID {
	case "empty":
		return nil, storage.ErrTokenNotFound
	default:
		return []storage.URLpairs{
			{
				OriginalURL: "url",
				ShortURL:    "shorturl",
			},
		}, nil
	}
}

func (st *Storage) GetURL(string) (string, error)                                 { return "https://ya.ru", nil }
func (st *Storage) GetTokensByUserID(string) ([]*tkn.Token, error)                { return nil, nil }
func (st *Storage) GetTokenByURL(string) (*tkn.Token, error)                      { return nil, nil }
func (st *Storage) HasURL(string) (bool, error)                                   { return true, nil }
func (st *Storage) Set(string, string, *tkn.Token) error                          { return nil }
func (st *Storage) SetBatch(userID string, url2token map[string]*tkn.Token) error { return nil }
func (st *Storage) Ping(ctx context.Context) error                                { return nil }
func (st *Storage) RemoveTokens(tokenValues []string, userID string) error        { return nil }

var us2 = &app.URLShortener{
	Config: &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	},
	Storage:        new(Storage),
	TokenGenerator: generator.NewSimple(),
}

func TestGet(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	type request struct {
		method string
		target string
		userID string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "success",
			request: request{
				method: http.MethodGet,
				target: "/qwerty",
				userID: "XXX-YYY-ZZZ",
			},
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "https://ya.ru",
				contentType: "",
			},
		},
		{
			name: "notfound",
			request: request{
				method: http.MethodGet,
				target: "/notfound",
				userID: "XXX-YYY-ZZZ",
			},
			want: want{
				code:        http.StatusNotFound,
				response:    "404 Not Found.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "expired",
			request: request{
				method: http.MethodGet,
				target: "/expired",
				userID: "XXX-YYY-ZZZ",
			},
			want: want{
				code:        498,
				response:    "498 Invalid Token.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "removed",
			request: request{
				method: http.MethodGet,
				target: "/removed",
				userID: "XXX-YYY-ZZZ",
			},
			want: want{
				code:        http.StatusGone,
				response:    "410 Gone.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "badrequest",
			request: request{
				method: http.MethodGet,
				target: "/",
				userID: "XXX-YYY-ZZZ",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "400 Bad Request.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, tt.request.target, nil)
			ctx := request.Context()
			ctx = context.WithValue(ctx, middleware.UserIDContextKey, tt.request.userID)
			request = request.WithContext(ctx)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Get(us2, w, r)
			})
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equalf(t, tt.want.response, string(resBody), w.Body.String())
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	type request struct {
		method string
		target string
		userID string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "success",
			request: request{
				method: http.MethodGet,
				target: "/qwerty",
				userID: "XXX-YYY-ZZZ",
			},
			want: want{
				code:        http.StatusOK,
				response:    `[{"original_url":"url", "short_url": "http://localhost:8080/shorturl"}]`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "empty",
			request: request{
				method: http.MethodGet,
				target: "/empty",
				userID: "empty",
			},
			want: want{
				code:        http.StatusNoContent,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, tt.request.target, nil)
			ctx := request.Context()
			ctx = context.WithValue(ctx, middleware.UserIDContextKey, tt.request.userID)
			request = request.WithContext(ctx)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				GetUserURLs(us2, w, r)
			})
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if helper.HasContentType(request, "application/json") {
				assert.JSONEqf(t, tt.want.response, string(resBody), w.Body.String())
			}
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func ExampleGet() {
	//nolint
	r, err := http.Get("http://localhost:8080/oTHlXx")
	if err != nil {
		fmt.Println("get error", err)
		return
	}
	defer r.Body.Close()

	fmt.Println(r.StatusCode)
	// 200
}

func ExampleGetUserURLs() {
	//nolint
	r, err := http.Get("http://localhost:8080/api/user/urls")
	if err != nil {
		fmt.Println("get error", err)
		return
	}
	defer r.Body.Close()

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}

	fmt.Println(string(buf))
	// [
	//  {"short_url": "http://localhost:8080/koRTZS", "original_url": "https://google.ru"}
	// ]
}
