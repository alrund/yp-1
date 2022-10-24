package grpcserver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/encryption"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
	pb "github.com/alrund/yp-1/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGet(t *testing.T) {
	testConfig := &config.Config{
		GrpcServerAddress: "localhost:9090",
	}
	testStorage := storage.NewMap()
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"https://ya.ru",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"https://ya.ru",
		&tkn.Token{Value: "expired", Expire: time.Now().Add(-tkn.LifeTime)},
	)
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"https://ya.ru",
		&tkn.Token{Value: "removed", Expire: time.Now().Add(tkn.LifeTime), Removed: true},
	)
	testTokenGenerator := new(TestGenerator)

	us := &app.URLShortener{
		Config:         testConfig,
		Storage:        testStorage,
		TokenGenerator: testTokenGenerator,
	}

	conn, err := grpc.DialContext(
		context.Background(),
		testConfig.GrpcServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer(us)),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewAppClient(conn)

	tests := []struct {
		name    string
		request *pb.GetRequest
		want    *pb.GetResponse
	}{
		{
			name:    "success",
			request: &pb.GetRequest{Token: "qwerty"},
			want: &pb.GetResponse{
				Message: http.StatusText(http.StatusTemporaryRedirect),
				Code:    http.StatusTemporaryRedirect,
				Url:     "https://ya.ru",
			},
		},
		{
			name:    "notfound",
			request: &pb.GetRequest{Token: "notfound"},
			want: &pb.GetResponse{
				Message: http.StatusText(http.StatusNotFound),
				Code:    http.StatusNotFound,
				Url:     "",
			},
		},
		{
			name:    "expired",
			request: &pb.GetRequest{Token: "expired"},
			want: &pb.GetResponse{
				Message: "498 Invalid Token.",
				Code:    498,
				Url:     "",
			},
		},
		{
			name:    "removed",
			request: &pb.GetRequest{Token: "removed"},
			want: &pb.GetResponse{
				Message: http.StatusText(http.StatusGone),
				Code:    http.StatusGone,
				Url:     "",
			},
		},
		{
			name:    "badrequest",
			request: &pb.GetRequest{Token: ""},
			want: &pb.GetResponse{
				Message: http.StatusText(http.StatusBadRequest),
				Code:    http.StatusBadRequest,
				Url:     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(context.Background(), tt.request)
			require.Nil(t, err)

			assert.Equal(t, tt.want.Message, resp.Message)
			assert.Equal(t, tt.want.Code, resp.Code)
			assert.Equal(t, tt.want.Url, resp.Url)
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	testConfig := &config.Config{
		GrpcServerAddress: "localhost:9090",
		BaseURL:           "http://localhost:8080/",
		CipherPass:        "PASS",
	}
	testStorage := storage.NewMap()
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"https://ya.ru",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)
	testTokenGenerator := new(TestGenerator)
	testEncryptor := encryption.NewEncryption(testConfig.CipherPass)

	us := &app.URLShortener{
		Config:         testConfig,
		Storage:        testStorage,
		TokenGenerator: testTokenGenerator,
	}

	conn, err := grpc.DialContext(
		context.Background(),
		testConfig.GrpcServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer(us)),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewAppClient(conn)

	type request struct {
		userID  string
		request *pb.GetUserURLsRequest
	}
	tests := []struct {
		name    string
		request request
		want    *pb.GetUserURLsResponse
	}{
		{
			name: "success",
			request: request{
				userID:  "XXX-YYY-ZZZ",
				request: &pb.GetUserURLsRequest{},
			},
			want: &pb.GetUserURLsResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
				Urls: []*pb.GetUserURLsResponse_Url{
					{
						OriginalUrl: "https://ya.ru",
						ShortUrl:    "http://localhost:8080/qwerty",
					},
				},
			},
		},
		{
			name: "notfound",
			request: request{
				userID:  "not-XXX-YYY-ZZZ",
				request: &pb.GetUserURLsRequest{},
			},
			want: &pb.GetUserURLsResponse{
				Message: http.StatusText(http.StatusNoContent),
				Code:    http.StatusNoContent,
				Urls:    []*pb.GetUserURLsResponse_Url(nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetUserURLs(getContextWithUserID(tt.request.userID, testEncryptor), tt.request.request)
			require.Nil(t, err)

			assert.Equal(t, tt.want.Message, resp.Message)
			assert.Equal(t, tt.want.Code, resp.Code)
			assert.Equal(t, tt.want.Urls, resp.Urls)
		})
	}
}
