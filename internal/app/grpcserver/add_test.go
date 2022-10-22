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

func TestAdd(t *testing.T) {
	testConfig := &config.Config{
		GrpcServerAddress: "localhost:9090",
		BaseURL:           "http://localhost:8080/",
	}
	testStorage := storage.NewMap()
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"existsurl",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)
	testTokenGenerator := new(TestGenerator)
	testToken, _ := testTokenGenerator.Generate()

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
		request *pb.AddRequest
		want    *pb.AddResponse
	}{
		{
			name:    "success",
			request: &pb.AddRequest{UserId: "XXX-YYY-ZZZ", Url: "http://ya.ru"},
			want: &pb.AddResponse{
				Error:     "",
				ErrorCode: http.StatusCreated,
				ShortUrl:  testConfig.BaseURL + testToken,
			},
		},
		{
			name:    "exists",
			request: &pb.AddRequest{UserId: "XXX-YYY-ZZZ", Url: "existsurl"},
			want: &pb.AddResponse{
				Error:     "",
				ErrorCode: http.StatusConflict,
				ShortUrl:  testConfig.BaseURL + testToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Add(context.Background(), tt.request)
			require.Nil(t, err)

			assert.Equal(t, tt.want.Error, resp.Error)
			assert.Equal(t, tt.want.ErrorCode, resp.ErrorCode)
			assert.Equal(t, tt.want.ShortUrl, resp.ShortUrl)
		})
	}
}
