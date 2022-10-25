package grpcserver

import (
	"context"
	"errors"

	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JSONRequest struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

// Add adds a URL string to shorten.
func (s *Server) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	var response pb.AddResponse

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		return &response, status.Error(codes.Internal, codes.Internal.String())
	}

	token, err := s.us.Add(userID, in.Url)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			return &response, status.Error(codes.Internal, err.Error())
		}
	}

	response.ShortUrl = s.us.GetBaseURL() + token.Value

	return &response, nil
}
