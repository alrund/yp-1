package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
	pb "github.com/alrund/yp-1/internal/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
		name     string
		config   *config.Config
		request  *pb.StatsRequest
		want     *pb.StatsResponse
		wantCode codes.Code
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
				Users: 1,
				Urls:  2,
			},
			wantCode: codes.OK,
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
				Users: 0,
				Urls:  0,
			},
			wantCode: codes.PermissionDenied,
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
				Users: 0,
				Urls:  0,
			},
			wantCode: codes.PermissionDenied,
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
				Users: 0,
				Urls:  0,
			},
			wantCode: codes.PermissionDenied,
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
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.Equal(t, tt.wantCode, e.Code())
				}
			} else {
				assert.Equal(t, tt.want.Users, resp.Users)
				assert.Equal(t, tt.want.Urls, resp.Urls)
			}
		})
	}
}
