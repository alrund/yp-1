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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestDeleteURLs(t *testing.T) {
	testConfig := &config.Config{
		GrpcServerAddress: "localhost:9090",
		BaseURL:           "http://localhost:8080/",
		CipherPass:        "PASS",
	}

	testTokenGenerator := new(TestGenerator)
	testEncryptor := encryption.NewEncryption(testConfig.CipherPass)

	type request struct {
		userID  string
		request *pb.DeleteURLsRequest
	}
	type want struct {
		response *pb.DeleteURLsResponse
		num      int
	}
	tests := []struct {
		name    string
		request *request
		want    want
	}{
		{
			name: "success",
			request: &request{
				userID:  "XXX-YYY-ZZZ",
				request: &pb.DeleteURLsRequest{Tokens: []*pb.DeleteURLsRequest_Token{{Value: "xxx"}}},
			},
			want: want{
				response: &pb.DeleteURLsResponse{},
				num:      1,
			},
		},
		{
			name: "success two",
			request: &request{
				userID:  "XXX-YYY-ZZZ",
				request: &pb.DeleteURLsRequest{Tokens: []*pb.DeleteURLsRequest_Token{{Value: "xxx"}, {Value: "yyy"}}},
			},
			want: want{
				response: &pb.DeleteURLsResponse{},
				num:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStorage := storage.NewMap()
			_ = testStorage.Set("XXX-YYY-ZZZ", "http://ya.ru", &tkn.Token{
				Value:  "xxx",
				Expire: time.Now().Add(tkn.LifeTime),
			})
			_ = testStorage.Set("XXX-YYY-ZZZ", "http://google.com", &tkn.Token{
				Value:  "yyy",
				Expire: time.Now().Add(tkn.LifeTime),
			})

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

			_, err = client.DeleteURLs(getContextWithUserID(tt.request.userID, testEncryptor), tt.request.request)
			require.Nil(t, err)

			if tt.want.num == 0 {
				return
			}

			time.Sleep(100 * time.Millisecond)

			tokens, err := testStorage.GetTokensByUserID(tt.request.userID)
			assert.NoError(t, err)
			var num int
			for _, token := range tokens {
				if token.Removed {
					continue
				}
				num++
			}

			assert.Equal(t, tt.want.num, num)
		})
	}
}
