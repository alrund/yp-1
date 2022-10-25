package grpcserver

import (
	"context"
	"errors"

	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AddBatch adds multiple URLs at once for shortening.
func (s *Server) AddBatch(ctx context.Context, in *pb.AddBatchRequest) (*pb.AddBatchResponse, error) {
	var response pb.AddBatchResponse

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		return &response, status.Error(codes.Internal, codes.Internal.String())
	}

	URLs, URL2Row := getURL2Row(in.Urls)
	tokens, err := s.us.AddBatch(userID, URLs)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			return &response, status.Error(codes.Internal, err.Error())
		}
	}

	for URL, token := range tokens {
		row, ok := URL2Row[URL]
		if !ok {
			return &response, status.Error(codes.Internal, "URL not found in URL2Row map")
		}
		response.ShortUrls = append(response.ShortUrls, &pb.AddBatchResponse_Url{
			CorrelationId: row.CorrelationId,
			ShortUrl:      s.us.GetBaseURL() + token.Value,
		})
	}

	return &response, nil
}

func getURL2Row(rows []*pb.AddBatchRequest_Url) ([]string, map[string]*pb.AddBatchRequest_Url) {
	URLs := make([]string, 0)
	URL2Row := map[string]*pb.AddBatchRequest_Url{}

	for _, row := range rows {
		URLs = append(URLs, row.OriginalUrl)
		URL2Row[row.OriginalUrl] = row
	}

	return URLs, URL2Row
}
