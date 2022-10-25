package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/encryption"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
	pb "github.com/alrund/yp-1/internal/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
		name     string
		request  *pb.GetRequest
		want     *pb.GetResponse
		wantCode codes.Code
	}{
		{
			name:    "success",
			request: &pb.GetRequest{Token: "qwerty"},
			want: &pb.GetResponse{
				Url: "https://ya.ru",
			},
			wantCode: codes.OK,
		},
		{
			name:    "notfound",
			request: &pb.GetRequest{Token: "notfound"},
			want: &pb.GetResponse{
				Url: "",
			},
			wantCode: codes.NotFound,
		},
		{
			name:    "expired",
			request: &pb.GetRequest{Token: "expired"},
			want: &pb.GetResponse{
				Url: "",
			},
			wantCode: codes.ResourceExhausted,
		},
		{
			name:    "removed",
			request: &pb.GetRequest{Token: "removed"},
			want: &pb.GetResponse{
				Url: "",
			},
			wantCode: codes.NotFound,
		},
		{
			name:    "badrequest",
			request: &pb.GetRequest{Token: ""},
			want: &pb.GetResponse{
				Url: "",
			},
			wantCode: codes.InvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(context.Background(), tt.request)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.Equal(t, tt.wantCode, e.Code())
				}
			} else {
				assert.Equal(t, tt.want.Url, resp.Url)
			}
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
		name     string
		request  request
		want     *pb.GetUserURLsResponse
		wantCode codes.Code
	}{
		{
			name: "success",
			request: request{
				userID:  "XXX-YYY-ZZZ",
				request: &pb.GetUserURLsRequest{},
			},
			want: &pb.GetUserURLsResponse{
				Urls: []*pb.GetUserURLsResponse_Url{
					{
						OriginalUrl: "https://ya.ru",
						ShortUrl:    "http://localhost:8080/qwerty",
					},
				},
			},
			wantCode: codes.OK,
		},
		{
			name: "notfound",
			request: request{
				userID:  "not-XXX-YYY-ZZZ",
				request: &pb.GetUserURLsRequest{},
			},
			want: &pb.GetUserURLsResponse{
				Urls: []*pb.GetUserURLsResponse_Url(nil),
			},
			wantCode: codes.NotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetUserURLs(getContextWithUserID(tt.request.userID, testEncryptor), tt.request.request)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.Equal(t, tt.wantCode, e.Code())
				}
			} else {
				assert.Equal(t, tt.want.Urls, resp.Urls)
			}
		})
	}
}
