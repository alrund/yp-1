package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
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
		method        string
		target        string
		userID        string
		errTypeUserID int
		body          string
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
		{
			name: "bad userID",
			request: request{
				method:        http.MethodPost,
				target:        "/",
				userID:        "",
				errTypeUserID: 666,
				body:          "url",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "500 Internal Server Error.\n",
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
			hc := NewCollection(us)
			request := getNewRequestWithUserID(
				tt.request.method,
				tt.request.target,
				tt.request.userID,
				tt.request.errTypeUserID,
				strings.NewReader(tt.request.body),
			)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(hc.Add())
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
		method        string
		target        string
		userID        string
		errTypeUserID int
		body          string
		contentType   string
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
				body:        `{"url": "https://ya1.ru"}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `{"result":"` + testConfig.BaseURL + testToken + `"}`,
				contentType: "application/json; charset=utf-8",
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
			hc := NewCollection(us)
			request := getNewRequestWithUserID(
				tt.request.method,
				tt.request.target,
				tt.request.userID,
				tt.request.errTypeUserID,
				strings.NewReader(tt.request.body),
			)
			request.Header.Set("Content-type", tt.request.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(hc.AddJSON())
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

func TestAddJSONFail(t *testing.T) {
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
		method        string
		target        string
		userID        string
		errTypeUserID int
		body          string
		contentType   string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "exists",
			request: request{
				method:      http.MethodPost,
				target:      "/api/shorten",
				userID:      "XXX-YYY-ZZZ",
				body:        `{"url": "existsurl"}`,
				contentType: "application/json; charset=utf-8",
			},
			want: want{
				code:        http.StatusConflict,
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
				body:        `{"url": "https://ya2.ru"}`,
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
				body:        `"url": "https://ya3.ru"}`,
				contentType: "application/json; charset=utf-8",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "400 Bad Request.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "bad userID",
			request: request{
				method:        http.MethodPost,
				target:        "/api/shorten",
				userID:        "",
				errTypeUserID: 666,
				body:          `{"url": "https://ya4.ru"}`,
				contentType:   "application/json; charset=utf-8",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "500 Internal Server Error.\n",
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
			hc := NewCollection(us)
			request := getNewRequestWithUserID(
				tt.request.method,
				tt.request.target,
				tt.request.userID,
				tt.request.errTypeUserID,
				strings.NewReader(tt.request.body),
			)
			request.Header.Set("Content-type", tt.request.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(hc.AddJSON())
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

func ExampleCollection_Add() {
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

func ExampleCollection_AddJSON() {
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
