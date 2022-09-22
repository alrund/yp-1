package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	"github.com/stretchr/testify/assert"
)

type TestGenerator struct{}

func (st *TestGenerator) Generate() (string, error) {
	return "qwerty", nil
}

func TestAdd(t *testing.T) {
	preparedStorage := storage.NewMap()
	_ = preparedStorage.Set(
		"XXX-YYY-ZZZ",
		"existsurl",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)

	testConfig := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	testTokenGenerator := new(TestGenerator)
	testToken, _ := testTokenGenerator.Generate()

	type want struct {
		code        int
		response    string
		contentType string
	}
	type request struct {
		method string
		target string
		userID string
		body   string
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
				userID: "XXX-YYY-ZZZ",
				body:   "url",
			},
			want: want{
				code:        http.StatusCreated,
				response:    testConfig.BaseURL + testToken,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "exists",
			request: request{
				method: http.MethodPost,
				target: "/",
				userID: "XXX-YYY-ZZZ",
				body:   "existsurl",
			},
			want: want{
				code:        http.StatusConflict,
				response:    testConfig.BaseURL + testToken,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &app.URLShortener{
				Config:         testConfig,
				Storage:        preparedStorage,
				TokenGenerator: testTokenGenerator,
			}

			request := getNewRequestWithUserID(
				tt.request.method,
				tt.request.target,
				tt.request.userID,
				strings.NewReader(tt.request.body),
			)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Add(us, w, r)
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
	testConfig := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	testTokenGenerator := new(TestGenerator)
	testToken, _ := testTokenGenerator.Generate()

	type want struct {
		code        int
		response    string
		contentType string
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
				method:      http.MethodPost,
				target:      "/api/shorten",
				userID:      "XXX-YYY-ZZZ",
				body:        `{"url": "https://ya.ru"}`,
				contentType: "application/json; charset=utf-8",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `{"result":"` + testConfig.BaseURL + testToken + `"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "success with content-type without charset",
			request: request{
				method:      http.MethodPost,
				target:      "/api/shorten",
				userID:      "XXX-YYY-ZZZ",
				body:        `{"url": "https://ya.ru"}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `{"result":"` + testConfig.BaseURL + testToken + `"}`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "incorrect content-type",
			request: request{
				method:      http.MethodPost,
				target:      "/api/shorten",
				userID:      "XXX-YYY-ZZZ",
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
			name: "bad request",
			request: request{
				method:      http.MethodPost,
				target:      "/api/shorten",
				userID:      "XXX-YYY-ZZZ",
				body:        `"url": "https://ya.ru"}`,
				contentType: "application/json; charset=utf-8",
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
			us := &app.URLShortener{
				Config:         testConfig,
				Storage:        storage.NewMap(),
				TokenGenerator: testTokenGenerator,
			}
			request := getNewRequestWithUserID(
				tt.request.method,
				tt.request.target,
				tt.request.userID,
				strings.NewReader(tt.request.body),
			)
			request.Header.Set("Content-type", tt.request.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				AddJSON(us, w, r)
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

func ExampleAdd() {
	//nolint
	r, err := http.Post(
		"http://localhost:8080/",
		"text/plain",
		bytes.NewBufferString("https://ya.ru"),
	)
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
	// "http://localhost:8080/oTHlXx"
}

func ExampleAddJSON() {
	//nolint
	r, err := http.Post(
		"http://localhost:8080/api/shorten",
		"application/json; charset=utf-8",
		strings.NewReader(`{"url": "https://ya.ru"}`),
	)
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
	// {"result":"http://localhost:8080/oTHlXx"}
}

func getNewRequestWithUserID(method, target, userID string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	ctx := request.Context()
	ctx = context.WithValue(ctx, middleware.UserIDContextKey, userID)
	return request.WithContext(ctx)
}
