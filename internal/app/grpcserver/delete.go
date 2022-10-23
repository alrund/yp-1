package grpcserver

import (
	"context"
	"net/http"

	pb "github.com/alrund/yp-1/internal/proto"
)

// DeleteURLs deletes shortened URL tokens.
func (s *Server) DeleteURLs(ctx context.Context, in *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	var response pb.DeleteURLsResponse

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	tokens := make([]string, 0)
	for _, t := range in.Tokens {
		tokens = append(tokens, t.GetValue())
	}

	go func() {
		_ = s.us.RemoveTokens(tokens, userID)
	}()

	response.ErrorCode = http.StatusAccepted
	response.Error = http.StatusText(http.StatusAccepted)

	return &response, nil
}
