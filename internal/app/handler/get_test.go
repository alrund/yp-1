package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
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
func (st *Storage) GetURL(string) (string, error)            { return "https://ya.ru", nil }
func (st *Storage) GetTokenByURL(string) (*tkn.Token, error) { return nil, nil }
func (st *Storage) HasURL(string) (bool, error)              { return true, nil }
func (st *Storage) Set(string, *tkn.Token) error             { return nil }

var us2 = &app.URLShortener{
	ServerAddress:  "localhost:8080",
	BaseURL:        "http://localhost:8080/",
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
			},
			want: want{
				code:        498,
				response:    "498 Invalid Token.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "incorrect method",
			request: request{
				method: http.MethodPost,
				target: "/",
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "Only GET requests are allowed!\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "incorrect path",
			request: request{
				method: http.MethodGet,
				target: "/",
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
