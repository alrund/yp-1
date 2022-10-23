package grpcserver

import (
	"context"
	"net/http"
	"testing"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPing(t *testing.T) {
	testConfig := &config.Config{
		GrpcServerAddress: "localhost:9090",
	}
	testStorage := storage.NewMap()
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
		request *pb.PingRequest
		want    *pb.PingResponse
	}{
		{
			name:    "success",
			request: &pb.PingRequest{},
			want: &pb.PingResponse{
				Error:     http.StatusText(http.StatusOK),
				ErrorCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Ping(context.Background(), tt.request)
			require.Nil(t, err)

			assert.Equal(t, tt.want.Error, resp.Error)
			assert.Equal(t, tt.want.ErrorCode, resp.ErrorCode)
		})
	}
}
