package app

import (
	"context"
	"errors"
	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
	"net/http"
)

type GRPCServer struct {
	pb.UnimplementedAppServer
	us *URLShortener
}

func NewGRPCServer(us *URLShortener) *GRPCServer {
	return &GRPCServer{us: us}
}

func (s *GRPCServer) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	var response pb.AddResponse

	token, err := s.us.Add(in.UserId, in.Url)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			response.ErrorCode = http.StatusInternalServerError
			response.Error = err.Error()
			return &response, nil
		}
		response.ErrorCode = http.StatusConflict
	}

	response.ShortUrl = token.Value

	return &response, nil
}
