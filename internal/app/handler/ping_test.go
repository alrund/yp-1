package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/stretchr/testify/assert"
)

var usPing = &app.URLShortener{
	Config: &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	},
	Storage:        new(TestStorage),
	TokenGenerator: generator.NewSimple(),
}

func TestPing(t *testing.T) {
	type want struct {
		code int
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
				target: "/ping",
				userID: "XXX-YYY-ZZZ",
			},
			want: want{
				code: http.StatusOK,
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
			h := http.HandlerFunc(Ping(usPing))
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
