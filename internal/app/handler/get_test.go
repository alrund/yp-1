package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

type Storage struct{}

func (st *Storage) HasToken(tokenValue string) (bool, error) {
	switch tokenValue {
	case "qwerty":
		return true, nil
	case "expired":
		return true, nil
	}
	return false, nil
}

func (st *Storage) GetToken(tokenValue string) (*tkn.Token, error) {
	if tokenValue == "expired" {
		return &tkn.Token{Value: "expired", Expire: time.Now().Add(-tkn.LifeTime)}, nil
	}
	return &tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)}, nil
}
func (st *Storage) GetURL(string) (string, error)                                 { return "https://ya.ru", nil }
func (st *Storage) GetTokensByUserID(string) ([]*tkn.Token, error)                { return nil, nil }
func (st *Storage) GetTokenByURL(string) (*tkn.Token, error)                      { return nil, nil }
func (st *Storage) GetURLsByUserID(string) ([]storage.URLpairs, error)            { return nil, nil }
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
