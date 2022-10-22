package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/stretchr/testify/assert"
)

type TestJSONResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ByCorrelationID []TestJSONResponse

func (a ByCorrelationID) Len() int           { return len(a) }
func (a ByCorrelationID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCorrelationID) Less(i, j int) bool { return a[i].CorrelationID < a[j].CorrelationID }

func TestAddBatchJSONSuccess(t *testing.T) {
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
				method: http.MethodPost,
				target: "/api/shorten/batch",
				userID: "XXX-YYY-ZZZ",
				body: `[
					{
						"correlation_id":"6d6bb7ef-78a5-49cd-a043-95233a79b54d",
						"original_url":"http://nxcfxrjohfr8.ru/aczlc5fcm5/tnypmcukjfip"
					},
					{
						"correlation_id":"591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
						"original_url":"http://rknawuufoxwpc.net/ejpjlw/qnulybd8720"
					}
				]`,
				contentType: "application/json; charset=utf-8",
			},
			want: want{
				code: http.StatusCreated,
				response: `[{"correlation_id":"591c1645-e1bb-4f64-bf8e-7eef7e5bff94","short_url":"` +
					testConfig.BaseURL +
					testToken +
					`"},{"correlation_id":"6d6bb7ef-78a5-49cd-a043-95233a79b54d","short_url":"` +
					testConfig.BaseURL +
					testToken +
					`"}]`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "success with content-type without charset",
			request: request{
				method: http.MethodPost,
				target: "/api/shorten/batch",
				userID: "XXX-YYY-ZZZ",
				body: `[
					{
						"correlation_id":"6d6bb7ef-78a5-49cd-a043-95233a79b54d",
						"original_url":"http://nxcfxrjohfr8.ru/aczlc5fcm5/tnypmcukjfip"
					},
					{
						"correlation_id":"591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
						"original_url":"http://rknawuufoxwpc.net/ejpjlw/qnulybd8720"
					}
				]`,
				contentType: "application/json",
			},
			want: want{
				code: http.StatusCreated,
				response: `[{"correlation_id":"6d6bb7ef-78a5-49cd-a043-95233a79b54d","short_url":"` +
					testConfig.BaseURL +
					testToken +
					`"},{"correlation_id":"591c1645-e1bb-4f64-bf8e-7eef7e5bff94","short_url":"` +
					testConfig.BaseURL +
					testToken +
					`"}]`,
				contentType: "application/json; charset=utf-8",
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
				tt.request.errTypeUserID,
				strings.NewReader(tt.request.body),
			)
			request.Header.Set("Content-type", tt.request.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(AddBatchJSON(us))
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			var wantResponses []TestJSONResponse
			err = json.Unmarshal([]byte(tt.want.response), &wantResponses)
			assert.NoError(t, err)

			var bodyResponses []TestJSONResponse
			err = json.Unmarshal(resBody, &bodyResponses)
			assert.NoError(t, err)

			sort.Sort(ByCorrelationID(wantResponses))
			sort.Sort(ByCorrelationID(bodyResponses))

			assert.Equal(t, wantResponses, bodyResponses)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestAddBatchJSONFail(t *testing.T) {
	testConfig := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	testTokenGenerator := new(TestGenerator)

	preparedStorage := storage.NewMap()
	_ = preparedStorage.Set(
		"XXX-YYY-ZZZ",
		"existsurl",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)

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
			name: "incorrect content-type",
			request: request{
				method: http.MethodPost,
				target: "/api/shorten/batch",
				userID: "XXX-YYY-ZZZ",
				body: `[
					{
						"correlation_id":"6d6bb7ef-78a5-49cd-a043-95233a79b54d",
						"original_url":"http://nxcfxrjohfr8.ru/aczlc5fcm5/tnypmcukjfip"
					},
					{
						"correlation_id":"591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
						"original_url":"http://rknawuufoxwpc.net/ejpjlw/qnulybd8720"
					}
				]`,
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
				target:      "/api/shorten/batch",
				userID:      "XXX-YYY-ZZZ",
				body:        `}`,
				contentType: "application/json",
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
				target:        "/api/shorten/batch",
				userID:        "",
				errTypeUserID: 666,
				body: `[
					{
						"correlation_id":"6d6bb7ef-78a5-49cd-a043-95233a79b54d",
						"original_url":"http://nxcfxrjohfr8.ru/aczlc5fcm5/tnypmcukjfip"
					},
					{
						"correlation_id":"591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
						"original_url":"http://rknawuufoxwpc.net/ejpjlw/qnulybd8720"
					}
				]`,
				contentType: "application/json",
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
			request := getNewRequestWithUserID(
				tt.request.method,
				tt.request.target,
				tt.request.userID,
				tt.request.errTypeUserID,
				strings.NewReader(tt.request.body),
			)
			request.Header.Set("Content-type", tt.request.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(AddBatchJSON(us))
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

func TestAddBatchJSONConflict(t *testing.T) {
	testConfig := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	testTokenGenerator := new(TestGenerator)
	testToken, _ := testTokenGenerator.Generate()

	preparedStorage := storage.NewMap()
	_ = preparedStorage.Set(
		"XXX-YYY-ZZZ",
		"existsurl",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)

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
			name: "conflict",
			request: request{
				method: http.MethodPost,
				target: "/api/shorten/batch",
				userID: "XXX-YYY-ZZZ",
				body: `[
					{
						"correlation_id":"222",
						"original_url":"existsurl"
					}
				]`,
				contentType: "application/json; charset=utf-8",
			},
			want: want{
				code: http.StatusConflict,
				response: `[{"correlation_id":"222","short_url":"` +
					testConfig.BaseURL +
					testToken +
					`"}]`,
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
			request := getNewRequestWithUserID(
				tt.request.method,
				tt.request.target,
				tt.request.userID,
				tt.request.errTypeUserID,
				strings.NewReader(tt.request.body),
			)
			request.Header.Set("Content-type", tt.request.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(AddBatchJSON(us))
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

func ExampleAddBatchJSON() {
	//nolint
	r, err := http.Post(
		"http://localhost:8080/api/shorten/batch",
		"application/json; charset=utf-8",
		strings.NewReader(`[
		{"correlation_id":"xxx","original_url":"https://ya.ru"},
		{"correlation_id":"yyy","original_url":"https://google.com"}
	]`),
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
	//[
	// {"correlation_id":"xxx","short_url":"http://localhost:8080/oTHlXx"},
	// {"correlation_id":"yyy","short_url":"http://localhost:8080/FaMvXd"}
	//]
}
