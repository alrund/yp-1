package grpcserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
)

func (s *Server) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	var response pb.AddResponse
	response.ErrorCode = http.StatusCreated

	token, err := s.us.Add(in.UserId, in.Url)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			response.ErrorCode = http.StatusInternalServerError
			response.Error = err.Error()
			return &response, nil
		}
		response.ErrorCode = http.StatusConflict
	}

	response.ShortUrl = s.us.GetBaseURL() + token.Value

	return &response, nil
}
