package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

type TestGenerator struct{}

func (st *TestGenerator) Generate() string {
	return "qwerty"
}

var us1 = &app.URLShortener{
	ServerAddress:  "localhost:8080",
	BaseURL:        "http://localhost:8080/",
	Storage:        storage.NewMap(),
	TokenGenerator: new(TestGenerator),
}

func TestAdd(t *testing.T) {
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
				method: http.MethodPost,
				target: "/",
			},
			want: want{
				code:        http.StatusCreated,
				response:    us1.GetBaseURL() + us1.TokenGenerator.Generate(),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "incorrect method",
			request: request{
				method: http.MethodGet,
				target: "/",
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "Only POST requests are allowed!\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "incorrect path",
			request: request{
				method: http.MethodPost,
				target: "/incorrect",
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
				Add(us1, w, r)
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

func TestAddJSON(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	type request struct {
		method      string
		target      string
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
				method:      http.MethodPost,
				target:      "/api/shorten",
				body:        `{"url": "https://ya.ru"}`,
				contentType: "application/json; charset=utf-8",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `{"result":"` + us1.GetBaseURL() + us1.TokenGenerator.Generate() + `"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "success with content-type without charset",
			request: request{
				method:      http.MethodPost,
				target:      "/api/shorten",
				body:        `{"url": "https://ya.ru"}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `{"result":"` + us1.GetBaseURL() + us1.TokenGenerator.Generate() + `"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "incorrect content-type",
			request: request{
				method:      http.MethodPost,
				target:      "/api/shorten",
				body:        `{"url": "https://ya.ru"}`,
				contentType: "text/plain; charset=utf-8",
			},
			want: want{
				code:        http.StatusUnsupportedMediaType,
				response:    "415 Unsupported Media Type.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "incorrect method",
			request: request{
				method: http.MethodGet,
				target: "/api/shorten",
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "Only POST requests are allowed!\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "incorrect path",
			request: request{
				method: http.MethodPost,
				target: "/incorrect",
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
			request := httptest.NewRequest(tt.request.method, tt.request.target, strings.NewReader(tt.request.body))
			request.Header.Set("Content-type", tt.request.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				AddJSON(us1, w, r)
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
