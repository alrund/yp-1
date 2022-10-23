package grpcserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
)

// Add adds a URL string to shorten.
func (s *Server) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	var response pb.AddResponse
	response.ErrorCode = http.StatusCreated

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	token, err := s.us.Add(userID, in.Url)
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
