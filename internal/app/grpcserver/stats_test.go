package grpcserver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
	pb "github.com/alrund/yp-1/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestStats(t *testing.T) {
	testStorage := storage.NewMap()
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"https://ya.ru",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"https://google.com",
		&tkn.Token{Value: "zxcvb", Expire: time.Now().Add(tkn.LifeTime)},
	)
	testTokenGenerator := new(TestGenerator)

	tests := []struct {
		name    string
		config  *config.Config
		request *pb.StatsRequest
		want    *pb.StatsResponse
	}{
		{
			name: "success",
			config: &config.Config{
				GrpcServerAddress: "localhost:9090",
				BaseURL:           "http://localhost:8080/",
				TrustedSubnet:     "216.58.192.64/24",
			},
			request: &pb.StatsRequest{XRealIP: "216.58.192.1"},
			want: &pb.StatsResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
				Users:   1,
				Urls:    2,
			},
		},
		{
			name: "fail - empty TrustedSubnet",
			config: &config.Config{
				GrpcServerAddress: "localhost:9090",
				BaseURL:           "http://localhost:8080/",
				TrustedSubnet:     "",
			},
			request: &pb.StatsRequest{XRealIP: "216.58.192.1"},
			want: &pb.StatsResponse{
				Message: http.StatusText(http.StatusForbidden),
				Code:    http.StatusForbidden,
				Users:   0,
				Urls:    0,
			},
		},
		{
			name: "fail - empty real ip",
			config: &config.Config{
				GrpcServerAddress: "localhost:9090",
				BaseURL:           "http://localhost:8080/",
				TrustedSubnet:     "216.58.192.64/24",
			},
			request: &pb.StatsRequest{XRealIP: ""},
			want: &pb.StatsResponse{
				Message: http.StatusText(http.StatusForbidden),
				Code:    http.StatusForbidden,
				Users:   0,
				Urls:    0,
			},
		},
		{
			name: "fail - not in TrustedSubnet",
			config: &config.Config{
				GrpcServerAddress: "localhost:9090",
				BaseURL:           "http://localhost:8080/",
				TrustedSubnet:     "216.58.192.64/24",
			},
			request: &pb.StatsRequest{XRealIP: "216.58.100.1"},
			want: &pb.StatsResponse{
				Message: http.StatusText(http.StatusForbidden),
				Code:    http.StatusForbidden,
				Users:   0,
				Urls:    0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &app.URLShortener{
				Config:         tt.config,
				Storage:        testStorage,
				TokenGenerator: testTokenGenerator,
			}

			conn, err := grpc.DialContext(
				context.Background(),
				tt.config.GrpcServerAddress,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialer(us)),
			)
			if err != nil {
				t.Fatalf("Failed to dial bufnet: %v", err)
			}
			defer conn.Close()
			client := pb.NewAppClient(conn)

			resp, err := client.Stats(context.Background(), tt.request)
			require.Nil(t, err)

			assert.Equal(t, tt.want.Message, resp.Message)
			assert.Equal(t, tt.want.Code, resp.Code)
			assert.Equal(t, tt.want.Users, resp.Users)
			assert.Equal(t, tt.want.Urls, resp.Urls)
		})
	}
}
