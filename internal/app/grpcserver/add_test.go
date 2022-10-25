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

func TestAdd(t *testing.T) {
	testConfig := &config.Config{
		GrpcServerAddress: "localhost:9090",
		BaseURL:           "http://localhost:8080/",
		CipherPass:        "PASS",
	}
	testStorage := storage.NewMap()
	_ = testStorage.Set(
		"XXX-YYY-ZZZ",
		"existsurl",
		&tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)},
	)
	testTokenGenerator := new(TestGenerator)
	testToken, _ := testTokenGenerator.Generate()
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
		request *pb.AddRequest
	}
	tests := []struct {
		name    string
		request *request
		want    *pb.AddResponse
	}{
		{
			name: "success",
			request: &request{
				userID:  "XXX-YYY-ZZZ",
				request: &pb.AddRequest{Url: "http://ya.ru"},
			},
			want: &pb.AddResponse{
				Message:  "",
				Code:     http.StatusCreated,
				ShortUrl: testConfig.BaseURL + testToken,
			},
		},
		{
			name: "exists",
			request: &request{
				userID:  "XXX-YYY-ZZZ",
				request: &pb.AddRequest{Url: "existsurl"},
			},
			want: &pb.AddResponse{
				Message:  "",
				Code:     http.StatusConflict,
				ShortUrl: testConfig.BaseURL + testToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Add(getContextWithUserID(tt.request.userID, testEncryptor), tt.request.request)
			require.Nil(t, err)

			assert.Equal(t, tt.want.Message, resp.Message)
			assert.Equal(t, tt.want.Code, resp.Code)
			assert.Equal(t, tt.want.ShortUrl, resp.ShortUrl)
		})
	}
}
