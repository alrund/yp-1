package grpcserver

import (
	"context"
	"errors"

	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token"
	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get returns a URL by token.
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	var response pb.GetResponse

	if in.Token == "" {
		return &response, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	url, err := s.us.Get(in.Token)
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			return &response, status.Error(codes.NotFound, codes.NotFound.String())
		}

		if errors.Is(err, token.ErrTokenExpiredError) {
			return &response, status.Error(codes.ResourceExhausted, codes.ResourceExhausted.String())
		}

		if errors.Is(err, token.ErrTokenRemovedError) {
			return &response, status.Error(codes.NotFound, codes.NotFound.String())
		}

		return &response, status.Error(codes.Internal, codes.Internal.String())
	}

	response.Url = url

	return &response, nil
}

// GetUserURLs returns a URL by user ID.
func (s *Server) GetUserURLs(ctx context.Context, in *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	var response pb.GetUserURLsResponse

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		return &response, status.Error(codes.Internal, codes.Internal.String())
	}

	urls, err := s.us.GetUserURLs(userID)
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			return &response, status.Error(codes.NotFound, codes.NotFound.String())
		}
		return &response, status.Error(codes.Internal, codes.Internal.String())
	}

	for _, url := range urls {
		response.Urls = append(response.Urls, &pb.GetUserURLsResponse_Url{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		})
	}

	return &response, nil
}
