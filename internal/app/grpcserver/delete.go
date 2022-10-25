package grpcserver

import (
	"context"

	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteURLs deletes shortened URL tokens.
func (s *Server) DeleteURLs(ctx context.Context, in *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	var response pb.DeleteURLsResponse

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		return &response, status.Error(codes.Internal, codes.Internal.String())
	}

	tokens := make([]string, 0)
	for _, t := range in.Tokens {
		tokens = append(tokens, t.GetValue())
	}

	go func() {
		_ = s.us.RemoveTokens(tokens, userID)
	}()

	return &response, nil
}
