package grpcserver

import (
	"context"
	"net/http"
	"sort"
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

type ByCorrelationID []*pb.AddBatchResponse_Url

func (a ByCorrelationID) Len() int           { return len(a) }
func (a ByCorrelationID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCorrelationID) Less(i, j int) bool { return a[i].CorrelationId < a[j].CorrelationId }

func TestAddBatchSuccess(t *testing.T) {
	testConfig := &config.Config{
		GrpcServerAddress: "localhost:9090",
		BaseURL:           "http://localhost:8080/",
		CipherPass:        "PASS",
	}
	testStorage := storage.NewMap()
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
		request *pb.AddBatchRequest
	}
	tests := []struct {
		name    string
		request *request
		want    *pb.AddBatchResponse
	}{
		{
			name: "success",
			request: &request{
				userID: "XXX-YYY-ZZZ",
				request: &pb.AddBatchRequest{
					Urls: []*pb.AddBatchRequest_Url{
						{
							CorrelationId: "6d6bb7ef-78a5-49cd-a043-95233a79b54d",
							OriginalUrl:   "http://nxcfxrjohfr8.ru/aczlc5fcm5/tnypmcukjfip",
						},
						{
							CorrelationId: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
							OriginalUrl:   "http://rknawuufoxwpc.net/ejpjlw/qnulybd8720",
						},
					},
				},
			},
			want: &pb.AddBatchResponse{
				Error:     "",
				ErrorCode: http.StatusCreated,
				ShortUrls: []*pb.AddBatchResponse_Url{
					{
						CorrelationId: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
						ShortUrl:      testConfig.BaseURL + testToken,
					},
					{
						CorrelationId: "6d6bb7ef-78a5-49cd-a043-95233a79b54d",
						ShortUrl:      testConfig.BaseURL + testToken,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.AddBatch(getContextWithUserID(tt.request.userID, testEncryptor), tt.request.request)
			require.Nil(t, err)

			sort.Sort(ByCorrelationID(tt.want.ShortUrls))
			sort.Sort(ByCorrelationID(resp.ShortUrls))

			assert.Equal(t, tt.want.Error, resp.Error)
			assert.Equal(t, tt.want.ErrorCode, resp.ErrorCode)
			assert.Equal(t, tt.want.ShortUrls, resp.ShortUrls)
		})
	}
}

func TestAddBatchFail(t *testing.T) {
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
		request *pb.AddBatchRequest
	}
	tests := []struct {
		name    string
		request *request
		want    *pb.AddBatchResponse
	}{
		{
			name: "exists",
			request: &request{
				userID: "",
				request: &pb.AddBatchRequest{
					Urls: []*pb.AddBatchRequest_Url{
						{
							CorrelationId: "6d6bb7ef-78a5-49cd-a043-95233a79b54d",
							OriginalUrl:   "existsurl",
						},
					},
				},
			},
			want: &pb.AddBatchResponse{
				Error:     "",
				ErrorCode: http.StatusConflict,
				ShortUrls: []*pb.AddBatchResponse_Url{
					{
						CorrelationId: "6d6bb7ef-78a5-49cd-a043-95233a79b54d",
						ShortUrl:      testConfig.BaseURL + testToken,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.AddBatch(getContextWithUserID(tt.request.userID, testEncryptor), tt.request.request)
			require.Nil(t, err)

			sort.Sort(ByCorrelationID(tt.want.ShortUrls))
			sort.Sort(ByCorrelationID(resp.ShortUrls))

			assert.Equal(t, tt.want.Error, resp.Error)
			assert.Equal(t, tt.want.ErrorCode, resp.ErrorCode)
			assert.Equal(t, tt.want.ShortUrls, resp.ShortUrls)
		})
	}
}
