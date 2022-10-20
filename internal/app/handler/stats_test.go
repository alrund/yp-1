package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/helper"
	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	type request struct {
		method  string
		target  string
		userID  string
		XRealIP string
	}
	tests := []struct {
		name    string
		config  *config.Config
		request request
		want    want
	}{
		{
			name: "success",
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080/",
				TrustedSubnet: "216.58.192.64/24",
			},
			request: request{
				method:  http.MethodGet,
				target:  "/api/internal/stats",
				userID:  "XXX-YYY-ZZZ",
				XRealIP: "216.58.192.1",
			},
			want: want{
				code:        http.StatusOK,
				response:    `[{"urls":"3", "users": "2"}]`,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "fail - empty TrustedSubnet",
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080/",
				TrustedSubnet: "",
			},
			request: request{
				method:  http.MethodGet,
				target:  "/api/internal/stats",
				userID:  "XXX-YYY-ZZZ",
				XRealIP: "",
			},
			want: want{
				code:        http.StatusForbidden,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "fail - empty real ip",
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080/",
				TrustedSubnet: "216.58.192.64/24",
			},
			request: request{
				method:  http.MethodGet,
				target:  "/api/internal/stats",
				userID:  "XXX-YYY-ZZZ",
				XRealIP: "",
			},
			want: want{
				code:        http.StatusForbidden,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "fail - not in TrustedSubnet",
			config: &config.Config{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080/",
				TrustedSubnet: "216.58.192.64/24",
			},
			request: request{
				method:  http.MethodGet,
				target:  "/api/internal/stats",
				userID:  "XXX-YYY-ZZZ",
				XRealIP: "216.58.100.1",
			},
			want: want{
				code:        http.StatusForbidden,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, tt.request.target, nil)
			request.Header.Add("X-Real-IP", tt.request.XRealIP)
			ctx := request.Context()
			ctx = context.WithValue(ctx, middleware.UserIDContextKey, tt.request.userID)
			request = request.WithContext(ctx)
			w := httptest.NewRecorder()

			usStats := &app.URLShortener{
				Config:         tt.config,
				Storage:        new(TestStorage),
				TokenGenerator: generator.NewSimple(),
			}

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Stats(usStats, w, r)
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
