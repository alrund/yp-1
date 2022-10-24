package grpcserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token"
	pb "github.com/alrund/yp-1/internal/proto"
)

// Get returns a URL by token.
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	var response pb.GetResponse

	if in.Token == "" {
		response.ErrorCode = http.StatusBadRequest
		response.Error = http.StatusText(http.StatusBadRequest)
		return &response, nil
	}

	url, err := s.us.Get(in.Token)
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			response.ErrorCode = http.StatusNotFound
			response.Error = http.StatusText(http.StatusNotFound)
			return &response, nil
		}

		if errors.Is(err, token.ErrTokenExpiredError) {
			response.ErrorCode = 498
			response.Error = "498 Invalid Token."
			return &response, nil
		}

		if errors.Is(err, token.ErrTokenRemovedError) {
			response.ErrorCode = http.StatusGone
			response.Error = http.StatusText(http.StatusGone)
			return &response, nil
		}

		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	response.ErrorCode = http.StatusTemporaryRedirect
	response.Error = http.StatusText(http.StatusTemporaryRedirect)
	response.Url = url

	return &response, nil
}

// GetUserURLs returns a URL by user ID.
func (s *Server) GetUserURLs(ctx context.Context, in *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	var response pb.GetUserURLsResponse

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	urls, err := s.us.GetUserURLs(userID)
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			response.ErrorCode = http.StatusNoContent
			response.Error = http.StatusText(http.StatusNoContent)
			return &response, nil
		}

		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	for _, url := range urls {
		response.Urls = append(response.Urls, &pb.GetUserURLsResponse_Url{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		})
	}
	response.ErrorCode = http.StatusOK
	response.Error = http.StatusText(http.StatusOK)

	return &response, nil
}
